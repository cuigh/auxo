package rpc

import (
	ct "context"
	"encoding/binary"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/data/guid"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/transport"
	"github.com/cuigh/auxo/util/retry"
)

type FailMode int

func (f *FailMode) Unmarshal(i interface{}) error {
	if s, ok := i.(string); ok {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "over":
			*f = FailOver
			return nil
		case "try":
			*f = FailTry
			return nil
		case "fast", "":
			*f = FailFast
			return nil
		}
	}
	return errors.Format("can't convert %v to FailMode", i)
}

const (
	// FailFast returns error immediately
	FailFast FailMode = iota
	// FailOver selects another server automatically
	FailOver
	// FailTry use current client again
	FailTry
	// FailBack records failed requests, resend in the future
	//FailBack
)

type nodeState int

const (
	// Idle indicates the Node is idle.
	stateIdle nodeState = iota
	//	// Connecting indicates the Node is connecting.
	//	Connecting
	// Ready indicates the Node is ready for work.
	stateReady
	//	// TransientFailure indicates the Node has seen a failure but expects to recover.
	//	TransientFailure
	// Shutdown indicates the Node has started shutting down.
	stateShutdown
)

var (
	manager = newClientManager()

	// ErrNodeUnavailable indicates no node is available for call.
	ErrNodeUnavailable = NewError(StatusNodeUnavailable, "rpc: no node is available")
	// ErrNodeShutdown indicates Node is shut down.
	ErrNodeShutdown = NewError(StatusNodeShutdown, "rpc: node is shut down")
)

//type ClientFilters struct {
//	BeforeDial func(n *Node)
//	AfterDial  func(n *Node)
//}

//type DialFilter func(DialHandler) DialHandler

//type DialHandler func(n *Node) error

type Identifier func() []byte

// Uint64 is an Identifier which generate an uint64 id.
func Uint64() Identifier {
	var n uint64
	return func() []byte {
		id := atomic.AddUint64(&n, 1)
		buf := make([]byte, 8)
		buf = buf[:binary.PutUvarint(buf, id)]
		return buf
		//return cast.StringToBytes(strconv.FormatUint(id, 10))
	}
}

// GUID is an Identifier using `guid.New`.
func GUID() []byte {
	return guid.New().Slice()
}

type NodeOptions struct {
	Codec struct {
		Name    string
		Options data.Map
	}
	Address transport.Address
	//ReadTimeout  time.Duration
	//WriteTimeout time.Duration
}

type Node struct {
	c *Client
	//opts    NodeOptions
	addr    transport.Address
	state   nodeState
	logger  *log.Logger
	handler CHandler
	calls   *callPool

	lock  sync.Mutex // protect dial
	ch    *Channel
	codec ClientCodec
}

func newNode(c *Client, addr transport.Address) *Node {
	n := &Node{
		c:      c,
		addr:   addr,
		logger: log.Get(PkgName),
		state:  stateIdle,
	}
	n.calls = newCallPool(n)
	return n
}

func (n *Node) initialize(ctx ct.Context) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state == stateReady {
		return nil
	}

	n.handler = n.do
	for i := len(n.c.filters) - 1; i >= 0; i-- {
		n.handler = n.c.filters[i](n.handler)
	}

	cb := codecs[n.c.opts.Codec.Name]
	if cb == nil {
		return NewError(StatusCodecNotRegistered, "rpc: codec '%s' not registered", n.c.opts.Codec.Name)
	}

	conn, err := transport.Dial(ctx, n.addr)
	if err != nil {
		return err
	}

	n.ch = newChannel(conn)
	n.codec = cb.NewClient(n.ch, n.c.opts.Codec.Options)
	go n.handle()
	n.state = stateReady
	return nil
}

//func (n *Node) State() nodeState {
//	return n.state
//}

//func (n *Node) Go(ctx ct.Context, service, method string, args []interface{}, reply interface{}) AsyncError {
//	c := n.calls.Acquire(n.c.id())
//	c.reset(ctx, service, method, args, reply)
//	err := n.handler(c)
//	err := n.send(c.req)
//	if err != nil {
//		return asyncError{err}
//	}
//	return c
//}

func (n *Node) Call(ctx ct.Context, service, method string, args []interface{}, reply interface{}) (err error) {
	c := n.calls.Acquire(n.c.id())
	c.reset(ctx, service, method, args, reply)
	defer n.calls.Release(c)
	err = n.handler(c)
	//n.calls.Release(c)
	return
}

// default handler
func (n *Node) do(c *Call) error {
	err := n.send(c.req)
	if err == nil {
		err = c.Wait()
	}
	return err
}

//func (n *Node) Login(ctx ct.Context, name, pwd string) error {
//	return n.Call(ctx, "$", "Login", nil, nil)
//}

func (n *Node) handle() {
	resp := &Response{}
	for n.state == stateReady {
		err := n.codec.DecodeHead(&resp.Head)
		if err != nil {
			if err == io.EOF {
				if n.state == stateShutdown {
					break
				} else {
					err = io.ErrUnexpectedEOF
				}
			}
			n.logger.Error("client > failed to decode head: ", err)
			n.Close()
			break
		}

		if len(resp.Head.ID) == 0 {
			n.heartbeat()
			continue
		}

		c := n.calls.Find(resp.Head.ID)
		if c == nil { // unknown response or request is timeout.
			n.codec.DiscardResult()
			n.logger.Error("client > unknown request: ", resp.Head.ID)
			continue
		}

		resp.Result.Value = c.reply
		err = n.codec.DecodeResult(&resp.Result)
		if err != nil {
			n.logger.Error("client > failed to decode result: ", err)
		} else if resp.Result.Error != nil {
			err = resp.Result.Error
		}
		c.finish(err)
	}

	n.calls.Clear(func(c *Call) {
		c.finish(ErrNodeShutdown)
	})
}

func (n *Node) send(req *Request) error {
	return n.codec.Encode(req)
}

func (n *Node) heartbeat() {
	n.codec.DiscardResult()
	err := n.send(&Request{})
	if err == nil {
		n.logger.Debug("client > heartbeat")
	} else {
		n.logger.Error("client > failed to send heartbeat response: ", err)
	}
}

func (n *Node) Close() {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state == stateReady {
		n.state = stateShutdown
		n.ch.Close()
		n.logger.Debug("client > node closed")
	}
}

type ClientOptions struct {
	Name    string
	Desc    string
	Version string
	Group   string
	Fail    FailMode
	Address []transport.Address
	Codec   struct {
		Name    string
		Options data.Map
	}
	Balancer struct {
		Name    string
		Options data.Map
	}
	Resolver struct {
		Name    string
		Options data.Map
	}
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (co *ClientOptions) AddAddress(uri string, options data.Map) {
	co.Address = append(co.Address, transport.Address{
		URL:     uri,
		Options: options,
	})
}

type Client struct {
	opts    ClientOptions
	logger  *log.Logger
	id      Identifier
	filters []CFilter

	lock     sync.Mutex
	addrs    []transport.Address
	nodes    []*Node
	resolver resolver.Resolver
	lb       Balancer
}

// NewClient creates Client with options.
func NewClient(opts ClientOptions) *Client {
	c := &Client{
		logger: log.Get(PkgName),
		opts:   opts,
		id:     Uint64(),
	}
	return c
}

// Dial creates Client with codec and addrs.
func Dial(codec string, addrs ...transport.Address) *Client {
	opts := ClientOptions{
		Address: addrs,
	}
	opts.Codec.Name = codec
	return NewClient(opts)
}

// AutoClient loads options from config file and create a Client. The created client is cached,
// so next call AutoClient with the same name will return the same Client instance.
func AutoClient(name string) (*Client, error) {
	return manager.Get(name)
}

func (c *Client) Use(filter ...CFilter) {
	c.filters = append(c.filters, filter...)
}

// todo: need to find a better way to handle retry when call with asynchronous.
//func (c *Client) Go(ctx ct.Context, service, method string, args []interface{}, reply interface{}) AsyncError {
//	n := c.lb.Next()
//	return n.Go(ctx, service, method, args, reply)
//}

func (c *Client) Call(ctx ct.Context, service, method string, args []interface{}, reply interface{}) error {
	n, err := c.getNode()
	if err != nil {
		return err
	}

	err = n.Call(ctx, service, method, args, reply)
	if err == nil || c.opts.Fail == FailFast || StatusOf(err) > 100 {
		return err
	}

	if c.opts.Fail == FailTry {
		// todo: allow customizing retry count
		return retry.Do(2, nil, func() error {
			return n.Call(ctx, service, method, args, reply)
		})
	} else if c.opts.Fail == FailOver {
		for _, n := range c.nodes {
			err = n.Call(ctx, service, method, args, reply)
			if err == nil || StatusOf(err) > 100 {
				return err
			}
		}
	}
	return err
}

func (c *Client) getNode() (n *Node, err error) {
	if c.nodes == nil {
		c.lock.Lock()
		defer c.lock.Unlock()

		if c.nodes == nil {
			err = c.init()
			if err != nil {
				return
			}
		}
	}

	n, err = c.lb.Next()
	if err != nil {
		return
	}

	if n.state != stateReady {
		ctx := ct.TODO()
		if c.opts.DialTimeout > 0 {
			var cancel ct.CancelFunc
			ctx, cancel = ct.WithTimeout(ct.TODO(), c.opts.DialTimeout)
			defer cancel()
		}
		err = n.initialize(ctx)
	}
	return
}

func (c *Client) Close() {
	for _, n := range c.nodes {
		n.Close()
	}
	c.nodes = nil

	if c.resolver != nil {
		c.resolver.Close()
		c.resolver = nil
	}
}

func (c *Client) init() error {
	if c.lb == nil {
		c.initBalancer()
	}

	if c.resolver == nil {
		err := c.initResolver()
		if err != nil {
			return err
		}
		c.resolver.Watch(c.notify)
	}

	addrs, err := c.resolver.Resolve()
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return ErrNodeUnavailable
	}

	c.updateNodes(addrs)
	return nil
}

func (c *Client) initBalancer() {
	var b BalancerBuilder
	if c.opts.Balancer.Name == "" {
		b = GetBalancer("random")
	} else {
		b = GetBalancer(c.opts.Balancer.Name)
		if b == nil {
			b = GetBalancer("random")
			c.logger.Warn("rpc > unknown balancer '%s', use 'random' instead", c.opts.Balancer.Name)
		}
	}
	c.lb = b.Build(c.opts.Balancer.Options)
}

func (c *Client) initResolver() error {
	name := c.opts.Resolver.Name
	if name == "" || name == "direct" {
		c.resolver = resolver.Direct(c.opts.Address...)
		return nil
	}

	if b := resolver.Get(name); b != nil {
		c.resolver = b.Build(resolver.Client{
			Server:  c.opts.Name,
			Version: c.opts.Version,
			Group:   c.opts.Group,
		}, c.opts.Resolver.Options)
		return nil
	}
	return errors.Format("rpc: unknown resolver '%s'", name)
}

func (c *Client) notify(addrs []transport.Address) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// prevent dropping all nodes
	if len(addrs) == 0 {
		return
	}

	var updated bool
	if len(addrs) == len(c.addrs) {
		sort.Slice(addrs, func(i, j int) bool { return addrs[i].URL < addrs[j].URL })
		for i, addr := range c.addrs {
			if addrs[i].URL != addr.URL {
				updated = true
				break
			}
		}
	} else {
		updated = true
	}

	if updated {
		c.updateNodes(addrs)
	}
}

func (c *Client) updateNodes(addrs []transport.Address) {
	addrMap := make(map[string]*transport.Address)
	for _, addr := range addrs {
		addrMap[addr.URL] = &addr
	}

	var nodes, invalid []*Node
	// keep the nodes still valid
	for _, n := range c.nodes {
		u := n.addr.URL
		if addr, ok := addrMap[u]; ok {
			n.addr.Options = addr.Options
			nodes = append(nodes, n)
			delete(addrMap, u)
		} else {
			invalid = append(invalid, n)
		}
	}
	// add new nodes
	for _, addr := range addrMap {
		nodes = append(nodes, newNode(c, *addr))
	}
	c.addrs = addrs
	c.nodes = nodes
	c.lb.Update(nodes)

	// close all invalid nodes
	for _, n := range invalid {
		n.Close()
	}
}

type clientManager struct {
	sync.RWMutex
	clients map[string]*Client
}

func newClientManager() *clientManager {
	return &clientManager{
		clients: make(map[string]*Client),
	}
}

func (m *clientManager) Get(name string) (c *Client, err error) {
	m.RLock()
	c = m.clients[name]
	m.RUnlock()
	if c != nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	c = m.clients[name]
	if c != nil {
		return
	}

	c, err = m.create(name)
	if err == nil {
		m.clients[name] = c
	}
	return
}

func (m *clientManager) create(name string) (c *Client, err error) {
	// todo: create from registry
	key := "rpc.client." + name
	if !config.Exist(key) {
		return nil, errors.Format("rpc: can't find config for client '%s'", name)
	}

	opts := ClientOptions{}
	err = config.UnmarshalOption(key, &opts)
	if err != nil {
		return nil, err
	}
	opts.Name = name
	return NewClient(opts), nil
}

package rpc

import (
	ct "context"
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"strings"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/data/guid"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
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

	//GlobalClientFilters ClientFilters
)

//type ClientFilters struct {
//	BeforeDial func(n *Node)
//	AfterDial  func(n *Node)
//	BeforeCall []CFilter
//	AfterCall  []CFilter
//}

//type DialFilter func(DialHandler) DialHandler

//type DialHandler func(n *Node) error

type Address struct {
	// URL is the server address on which a connection will be established.
	URL string
	// Codec is the codec name of this address.
	//Codec string
	// Options is the information associated with Address, which may be used
	// to make load balancing decision.
	Options data.Map
}

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

//// CallOption configures a Call before it starts or extracts information from
//// a Call after it completes.
//type CallOption interface {
//	// before is called before the call is sent to any server.  If before
//	// returns a non-nil error, the RPC fails with that error.
//	before(*call) error
//
//	// after is called after the call has completed.  after cannot return an
//	// error, so any failures should be reported via output parameters.
//	after(*call)
//}

type NodeOptions struct {
	Codec struct {
		Name    string
		Options data.Map
	}
	Address Address
	//ReadTimeout  time.Duration
	//WriteTimeout time.Duration
}

type Node struct {
	c       *Client
	opts    NodeOptions
	state   nodeState
	logger  *log.Logger
	handler CHandler
	calls   *callPool

	lock  sync.Mutex // protect dial
	ch    *Channel
	codec ClientCodec
}

func newNode(c *Client, opts NodeOptions) *Node {
	n := &Node{
		c:      c,
		opts:   opts,
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

	cb := codecs[n.opts.Codec.Name]
	if cb == nil {
		return NewError(StatusCodecNotRegistered, "rpc: codec '%s' not registered", n.opts.Codec.Name)
	}

	conn, err := transport.Dial(ctx, n.opts.Address.URL, n.opts.Address.Options)
	if err != nil {
		return err
	}

	n.ch = newChannel(conn)
	n.codec = cb.NewClient(n.ch, n.opts.Codec.Options)
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
	//Address   []string // todo: etcd://, swarm://, tcp://?
	Address  []Address
	Balancer string
	//Balancer struct{
	//	Name string
	//	Options data.Map
	//}
	Codec struct {
		Name    string
		Options data.Map
	}
	Resolver struct {
		Name    string
		Options data.Map
	}
	Registry struct {
		Name    string
		Options data.Map
	}
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (co *ClientOptions) AddAddress(uri string, options data.Map) {
	co.Address = append(co.Address, Address{
		URL:     uri,
		Options: options,
	})
}

type Client struct {
	opts    ClientOptions
	logger  *log.Logger
	id      Identifier
	nodes   []*Node
	lb      Balancer
	fail    FailMode
	retry   int32
	filters []CFilter
}

// NewClient creates Client with options.
func NewClient(opts ClientOptions) *Client {
	c := &Client{
		logger: log.Get(PkgName),
		opts:   opts,
		id:     Uint64(),
	}
	for _, addr := range c.opts.Address {
		if addr.Options.Get("codec") == "" {
			addr.Options["codec"] = c.opts.Codec.Name
		}
		nodeOpts := NodeOptions{
			Address: addr,
			Codec:   opts.Codec,
		}
		n := newNode(c, nodeOpts)
		c.nodes = append(c.nodes, n)
	}
	c.initBalancer()
	c.lb.Update(c.nodes)
	return c
}

// Dial creates Client with codec and addrs.
func Dial(codec string, addrs ...Address) *Client {
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

//func (c *Client) notify(addrs []*Address) {
//
//}

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
	if err == nil || c.fail == FailFast || StatusOf(err) > 100 {
		return err
	}

	if c.fail == FailTry {
		// todo: allow customizing retry count
		return retry.Do(2, nil, func() error {
			return n.Call(ctx, service, method, args, reply)
		})
	} else if c.fail == FailOver {
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
}

func (c *Client) initBalancer() {
	b := GetBalancer(c.opts.Balancer)
	if b == nil {
		c.logger.Warn("client > use default balancer: random")
		b = GetBalancer("random")
	}
	c.lb = b.Build(data.Map{})
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

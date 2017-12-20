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
	"github.com/cuigh/auxo/util/cast"
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
//	BeforeDial   func(n *Node)
//	AfterDial    func(n *Node)
//	BeforeSend   func(ctx ct.Context, req *Request)
//	AfterReceive func(resp *Response)
//}

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

type AsyncError interface {
	Wait() error
}

type asyncError struct {
	error
}

func (ae asyncError) Wait() error {
	return ae.error
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

type call struct {
	n *Node
	Request
	ctx   ct.Context
	reply interface{}
	err   chan error
	//done  data.Chan
}

func (c *call) finish(err error) {
	c.err <- err
	//c.done.TrySend()
}

// Wait implements interface of AsyncError.
func (c *call) Wait() (err error) {
	ctx, cancel := ct.WithTimeout(c.ctx, time.Second*10)
	defer cancel()

	select {
	case err = <-c.err:
	case <-ctx.Done():
		c.n.pending.remove(c)
		if err = ctx.Err(); err == ct.Canceled {
			err = NewError(StatusCanceled, err.Error())
		} else {
			err = NewError(StatusDeadlineExceeded, err.Error())
		}
	}
	c.n.calls.put(c)
	return
}

func (c *call) reset(ctx ct.Context, id []byte, service, method string, args []interface{}, reply interface{}) {
	//c.Head.Type = 0
	c.Head.ID = id
	c.Head.Service = service
	c.Head.Method = method
	c.Args = args
	c.ctx = ctx
	c.reply = reply
	c.err = make(chan error, 1)
	//c.err = nil
	//c.done = make(data.Chan, 1)
}

type callPool struct {
	p sync.Pool
}

func (cp *callPool) get() *call {
	return cp.p.Get().(*call)
}

func (cp *callPool) put(c *call) {
	cp.p.Put(c)
}

type callMap struct {
	sync.Mutex
	m map[string]*call
}

func (m *callMap) get(id []byte) (c *call) {
	key := cast.BytesToString(id)
	m.Lock()
	c = m.m[key]
	if c != nil {
		delete(m.m, key)
	}
	m.Unlock()
	return
}

func (m *callMap) put(c *call) {
	m.Lock()
	m.m[cast.BytesToString(c.Head.ID)] = c
	m.Unlock()
}

func (m *callMap) remove(c *call) {
	m.Lock()
	delete(m.m, cast.BytesToString(c.Head.ID))
	m.Unlock()
}

func (m *callMap) clear(fn func(c *call)) {
	m.Lock()
	for k, c := range m.m {
		fn(c)
		delete(m.m, k)
	}
	m.Unlock()
}

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
	lock    sync.Mutex // protect dial
	opts    NodeOptions
	state   nodeState
	logger  *log.Logger
	id      Identifier
	ch      *Channel
	codec   ClientCodec
	pending callMap
	calls   callPool
}

func newNode(opts NodeOptions) *Node {
	n := &Node{
		opts:   opts,
		logger: log.Get(PkgName),
		//id:     Uint64(),
		state: stateIdle,
	}
	n.pending.m = make(map[string]*call)
	n.calls.p.New = func() interface{} {
		return &call{n: n}
	}
	return n
}

func (n *Node) dial(ctx ct.Context) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.state == stateReady {
		return nil
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

func (n *Node) Go(ctx ct.Context, service, method string, args []interface{}, reply interface{}) AsyncError {
	c := n.calls.get()
	c.reset(ctx, n.id(), service, method, args, reply)
	n.pending.put(c)

	err := n.send(&c.Request)
	if err != nil {
		n.pending.remove(c)
		n.calls.put(c)
		return asyncError{err}
	}
	return c
}

func (n *Node) Call(ctx ct.Context, service, method string, args []interface{}, reply interface{}) (err error) {
	c := n.calls.get()
	c.reset(ctx, n.id(), service, method, args, reply)
	n.pending.put(c)

	err = n.send(&c.Request)
	if err == nil {
		err = c.Wait()
	} else {
		n.pending.remove(c)
		n.calls.put(c)
	}
	return
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

		c := n.pending.get(resp.Head.ID)
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

	n.pending.clear(func(c *call) {
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
	opts   ClientOptions
	logger *log.Logger
	id     Identifier
	nodes  []*Node
	lb     Balancer
	fail   FailMode
	retry  int32
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
		n := newNode(nodeOpts)
		n.id = c.id
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

func (c *Client) notify(addrs []*Address) {

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
	if err == nil || c.errorCode(err) > 100 || c.fail == FailFast {
		return err
	}

	if c.fail == FailTry {
		// todo: retry count
		return retry.Do(2, nil, func() error {
			return n.Call(ctx, service, method, args, reply)
		})
	} else if c.fail == FailOver {
		for _, n := range c.nodes {
			err = n.Call(ctx, service, method, args, reply)
			if err == nil || c.errorCode(err) > 100 {
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
		err = n.dial(ctx)
	}
	return
}

func (c *Client) errorCode(err error) int32 {
	if e, ok := err.(*errors.CodedError); ok {
		return e.Code
	}
	return 0
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

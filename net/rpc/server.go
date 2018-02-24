package rpc

import (
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/times"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/net/transport"
	"github.com/cuigh/auxo/util/debug"
	"github.com/cuigh/auxo/util/run"
)

type matchInfo struct {
	matcher Matcher
	codec   string
	opts    data.Map
}

type Stats struct {
	AcceptSuccess uint32
	AcceptFailure uint32
}

type ServerOptions struct {
	Name     string `json:"name" yaml:"name"`
	Desc     string `json:"desc" yaml:"desc"`
	Version  string `json:"version" yaml:"version"`
	Address  []transport.Address
	Registry struct {
		Name    string   `json:"name" yaml:"name"`
		Options data.Map `json:"options" yaml:"options"`
	} `json:"registry" yaml:"registry"`
	ReadTimeout       time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout" yaml:"write_timeout"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval" yaml:"heartbeat"`
	MaxClients        int32         `json:"max_clients" yaml:"max_clients"`
	MaxPoolSize       int32         `json:"max_pool_size" yaml:"max_pool_size"`
	Backlog           int32         `json:"backlog" yaml:"backlog"`
}

func (opts *ServerOptions) ensure() error {
	if len(opts.Address) == 0 {
		return errors.New("rpc: address must be set for server")
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = 10 * time.Second
	}
	if opts.WriteTimeout <= 0 {
		opts.WriteTimeout = 10 * time.Second
	}
	if opts.MaxPoolSize <= 0 {
		opts.MaxPoolSize = 1024
	}
	if opts.MaxClients <= 0 {
		opts.MaxClients = 10000
	}
	if opts.MaxPoolSize <= 0 {
		opts.MaxPoolSize = 1024
	}
	if opts.Backlog <= 0 {
		opts.Backlog = 10000
	}
	return nil
}

func (opts *ServerOptions) AddAddress(uri string, options data.Map) {
	opts.Address = append(opts.Address, transport.Address{
		URL:     uri,
		Options: options,
	})
}

type Server struct {
	opts      ServerOptions
	logger    log.Logger
	matchers  []matchInfo
	registry  registry.Registry
	ctxPool   *contextPool
	filters   []SFilter
	listeners []net.Listener
	sessions  *sessionMap
	actions   *actionSet
	pool      *run.Pool
	hb        *times.Wheel // for heartbeat
	listening int32
}

func NewServer(opts ServerOptions) (*Server, error) {
	err := opts.ensure()
	if err != nil {
		return nil, err
	}

	s := &Server{
		logger:   log.Get(PkgName),
		opts:     opts,
		sessions: newSessionMap(),
		actions:  newActionSet(),
		pool:     &run.Pool{Max: opts.MaxPoolSize, Backlog: int(opts.Backlog)},
	}
	s.ctxPool = newContextPool(s)
	return s, nil
}

func Listen(addrs ...transport.Address) (*Server, error) {
	opts := ServerOptions{
		Address: addrs,
	}
	return NewServer(opts)
}

// AutoServer loads options from config file and create a Server.
func AutoServer(name string) (*Server, error) {
	key := "rpc.server." + name
	if !config.Exist(key) {
		return nil, errors.Format("rpc: can't find config for server '%s'", name)
	}

	opts := ServerOptions{}
	err := config.UnmarshalOption(key, &opts)
	if err != nil {
		return nil, err
	}
	opts.Name = name
	return NewServer(opts)
}

func (s *Server) Sessions() SessionMap {
	return s.sessions
}

func (s *Server) initRegistry() {
	if s.opts.Registry.Name == "" {
		return
	}

	b := registry.Get(s.opts.Registry.Name)
	if b == nil {
		s.logger.Warnf("rpc > Unknown registry '%v'", s.opts.Registry.Name)
		return
	}

	codecs := make([]string, len(s.matchers))
	for i, m := range s.matchers {
		codecs[i] = m.codec
	}

	var err error
	s.registry, err = b.Build(registry.Server{
		Name:      s.opts.Name,
		Version:   s.opts.Version,
		Addresses: s.opts.Address,
		Options: data.Map{
			"desc":        s.opts.Desc,
			"max_clients": s.opts.MaxClients,
			"codec":       strings.Join(codecs, ","),
		},
	}, s.opts.Registry.Options)
	if err == nil {
		s.registry.Register()
	} else {
		s.logger.Warnf("rpc > Failed to create registry '%v': %s", s.opts.Registry.Name, err)
	}
}

// todo
//func (s *Server) ErrorHandler()  {
//}

func (s *Server) Serve() error {
	if !atomic.CompareAndSwapInt32(&s.listening, 0, 1) {
		return errors.New("rpc: server is already running")
	}

	err := s.initListeners()
	if err != nil {
		return err
	}

	s.pool.Start()
	if s.opts.HeartbeatInterval > 0 {
		s.hb = times.NewWheel(time.Second, int(s.opts.HeartbeatInterval.Seconds()))
	}

	// todo: use errgroup.Group
	errs := make(chan error, len(s.listeners))
	for _, l := range s.listeners {
		go func(l net.Listener) {
			s.logger.Infof("rpc > Server is listening on %v", l.Addr())
			errs <- s.serve(l)
		}(l)
	}

	s.initRegistry()

	err = <-errs
	if err != ErrServerClosed {
		s.logger.Errorf("rpc > Failed to run server: %s", err)
		s.Close(0)
		s.listening = 0
	}
	return err
}

func (s *Server) initListeners() (err error) {
	s.listeners = make([]net.Listener, len(s.opts.Address))
	for i, addr := range s.opts.Address {
		s.listeners[i], err = transport.Listen(addr)
		if err != nil {
			break
		}
	}
	if err != nil {
		for _, l := range s.listeners {
			if l == nil {
				break
			}
			l.Close()
		}
	}
	return
}

func (s *Server) Close(timeout time.Duration) {
	if !atomic.CompareAndSwapInt32(&s.listening, 1, 0) {
		return
	}

	if s.registry != nil {
		s.registry.Close()
		s.registry = nil
	}

	for _, l := range s.listeners {
		l.Close()
	}

	if timeout > 0 {
		s.logger.Info("rpc > Try to close server...")
		if err := s.pool.Wait(timeout); err == nil {
			s.logger.Info("rpc > Server closed gracefully")
		} else {
			s.logger.Warn("rpc > Server closed with error: ", err.Error())
		}
	} else {
		s.logger.Info("rpc > Server closed")
	}
	s.sessions.Close()
	s.pool.Stop()
}

func (s *Server) Match(m Matcher, codec string, opts ...data.Map) {
	mi := matchInfo{matcher: m, codec: codec}
	if len(opts) > 0 {
		mi.opts = opts[0]
	}
	s.matchers = append(s.matchers, mi)
}

func (s *Server) Use(filter ...SFilter) {
	s.filters = append(s.filters, filter...)
}

func (s *Server) RegisterService(name string, svc interface{}, filter ...SFilter) error {
	return s.actions.RegisterService(name, svc, filter...)
}

func (s *Server) RegisterFunc(service, method string, fn interface{}, filter ...SFilter) error {
	return s.actions.RegisterFunc(service, method, fn, filter...)
}

func (s *Server) serve(l net.Listener) (err error) {
	var conn net.Conn
	for {
		conn, err = l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				// todo:
				time.Sleep(time.Millisecond * 100)
				continue
			}
			break
		}

		// close connection when max clients was reached.
		if s.sessions.Count() >= int(s.opts.MaxClients) {
			s.logger.Warnf("server > reach max clients, close connection: %v", conn.RemoteAddr())
			conn.Close()
			continue
		}

		ch := newChannel(conn)
		c := s.findCodec(ch)
		if c == nil {
			s.logger.Error("codec not found")
			conn.Close()
			continue
		}

		go s.handleSession(ch, c)
	}

	if s.listening == 0 {
		err = ErrServerClosed
	}
	return
}

func (s *Server) handleSession(ch *Channel, sc ServerCodec) {
	sn := s.addSession(ch, sc)
	defer func() {
		s.sessions.Remove(sn)
		ch.Close()
		if e := recover(); e != nil {
			s.logger.Errorf("server > failed to handle session: %v, stack: %s", e, debug.StackSkip(1))
		}
	}()

	for {
		ctx := s.ctxPool.Get(ch, sc)
		err := sc.DecodeHead(&ctx.req.Head)
		if err != nil {
			if err == io.EOF {
				s.logger.Debug("server > session closed")
			} else {
				s.logger.Errorf("server > decode head failed: %v", err)
			}
			//ch.Close()
			return
		}

		// heartbeat response
		if ctx.req.Head.ID == 0 {
			sc.DiscardArgs()
			sn.heartbeat = time.Now()
			continue
		}

		// If server is closing, send ErrServerClosed to client immediately.
		if s.listening == 0 {
			s.encode(ctx, nil, ErrServerClosed)
			continue
		}

		err = s.decodeArgs(sc, ctx)
		if err == nil {
			s.handleRequest(ctx, sc)
		} else {
			s.encode(ctx, nil, err)
		}
	}
}

func (s *Server) addSession(ch *Channel, sc ServerCodec) *session {
	sn := newSession(ch, sc)
	s.sessions.Add(sn)
	if s.opts.HeartbeatInterval > 0 {
		s.hb.Add(func() bool { return s.heartbeat(sn) })
	}
	return sn
}

func (s *Server) heartbeat(sn *session) bool {
	if s.sessions.Get(sn.id) != nil {
		if sn.heartbeat.Add(s.opts.HeartbeatInterval).Add(time.Second).Before(time.Now()) {
			s.logger.Info("server > close session [%v] for heartbeat timeout", sn.id)
			sn.Close()
			return false
		} else {
			err := sn.Encode(&Response{})
			if err != nil {
				s.logger.Error("server > failed to send heartbeat request: ", err)
			}
			return true
		}
	}
	return false
}

func (s *Server) handleRequest(ctx *context, sc ServerCodec) {
	err := s.pool.Put(func() {
		defer func() {
			if e := recover(); e != nil {
				s.logger.Error("server > failed to handle request: ", e)
			}
		}()

		// todo: move to initialization stage
		h := ctx.action.Handler()
		for i := len(s.filters) - 1; i >= 0; i-- {
			h = s.filters[i](h)
		}

		r, err := h(ctx)
		s.encode(ctx, r, err)
	})
	if err != nil {
		s.encode(ctx, nil, err)
	}
}

func (s *Server) decodeArgs(sc ServerCodec, ctx *context) (err error) {
	ctx.action = s.actions.Find(ctx.req.Head.Service, ctx.req.Head.Method)
	if ctx.action == nil {
		sc.DiscardArgs()
		return NewError(StatusMethodNotFound, "method not found: %s.%s", ctx.req.Head.Service, ctx.req.Head.Method)
	}

	args := ctx.action.fillArgs(ctx)
	return sc.DecodeArgs(args)
}

func (s *Server) encode(ctx *context, r interface{}, err error) {
	ctx.resp.Head.ID = ctx.req.Head.ID
	//ctx.resp.Head.Type = 0
	if err == nil {
		ctx.resp.Result.Value = r
	} else {
		ctx.resp.Result.Error = s.wrapError(err)
	}

	err = ctx.codec.Encode(ctx.resp)
	s.ctxPool.Put(ctx)
	if err != nil {
		s.logger.Error("encode response failed: ", err)
	}
}

func (s *Server) findCodec(ch *Channel) ServerCodec {
	for _, m := range s.matchers {
		if m.matcher == nil || m.matcher(ch) {
			if cb := codecs[m.codec]; cb != nil {
				return cb.NewServer(ch, m.opts)
			}
		}
	}
	return nil
}

func (s *Server) wrapError(err error) *errors.CodedError {
	if ce, ok := err.(*errors.CodedError); ok {
		return ce
	}
	return NewError(StatusUnknown, err.Error())
}

type Session interface {
	ID() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type session struct {
	*Channel
	ServerCodec
	heartbeat time.Time // heartbeat time
}

func newSession(ch *Channel, sc ServerCodec) *session {
	return &session{
		Channel:     ch,
		ServerCodec: sc,
		heartbeat:   time.Now(),
	}
}

type SessionMap interface {
	Range(func(s Session) bool)
}

func newSessionMap() *sessionMap {
	return &sessionMap{
		channels: make(map[string]*session),
	}
}

type sessionMap struct {
	lock     sync.RWMutex
	channels map[string]*session
}

func (m *sessionMap) Add(c *session) {
	m.lock.Lock()
	m.channels[c.ID()] = c
	m.lock.Unlock()
}

func (m *sessionMap) Get(id string) (s *session) {
	m.lock.RLock()
	s = m.channels[id]
	m.lock.RUnlock()
	return
}

func (m *sessionMap) Remove(c *session) {
	if c != nil {
		m.lock.Lock()
		delete(m.channels, c.ID())
		m.lock.Unlock()
	}
}

func (m *sessionMap) Close() {
	m.lock.Lock()
	for _, c := range m.channels {
		c.Close()
	}
	m.lock.Unlock()
}

func (m *sessionMap) Count() int {
	m.lock.RLock()
	c := len(m.channels)
	m.lock.RUnlock()
	return c
}

func (m *sessionMap) Range(fn func(s Session) bool) {
	m.lock.Lock()
	for _, c := range m.channels {
		if !fn(c) {
			break
		}
	}
	m.lock.Unlock()
}

package web

import (
	scontext "context"
	"crypto/tls"
	"fmt"
	slog "log"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web/router"
	"golang.org/x/crypto/acme/autocert"
)

var _ Router = &Server{}

// Server represents information of the HTTP server.
type Server struct {
	ErrorHandler ErrorHandler
	Binder       Binder
	Validator    Validator
	Renderer     Renderer
	Logger       log.Logger
	stdLogger    *slog.Logger
	cfg          *Options
	filters      []Filter
	router       *router.Tree
	routes       map[string]router.Route
	ctxPool      *contextPool
	servers      []*http.Server
}

// Default creates an instance of Server with default options.
func Default(address ...string) (server *Server) {
	c := &Options{}
	for _, addr := range address {
		c.Entries = append(c.Entries, Entry{Address: addr})
	}
	return New(c)
}

// Auto creates an instance of Server with options loaded from app.yaml/app.toml.
func Auto() (server *Server) {
	c := &Options{}
	err := config.UnmarshalOption("web", c)
	if err != nil {
		panic(err)
	}
	return New(c)
}

// New creates an instance of Server.
func New(c *Options) (s *Server) {
	c.ensure()
	s = &Server{
		cfg:    c,
		Logger: log.Get(PkgName),
		Binder: new(binder),
		router: router.New(router.Options{}),
		routes: make(map[string]router.Route),
	}
	s.stdLogger = slog.New(s.Logger, "web > ", 0)
	s.ctxPool = newContextPool(s)
	return s
}

// Router returns router.
func (s *Server) Router() *router.Tree {
	return s.router
}

// Use adds global filters to the router.
func (s *Server) Use(filters ...Filter) {
	s.filters = append(s.filters, filters...)
}

// UseFunc adds global filters to the router.
func (s *Server) UseFunc(filters ...FilterFunc) {
	for _, f := range filters {
		s.filters = append(s.filters, f)
	}
}

// Connect registers a route that matches 'CONNECT' method.
func (s *Server) Connect(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodConnect, path, h, opts...)
}

// Delete registers a route that matches 'DELETE' method.
func (s *Server) Delete(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodDelete, path, h, opts...)
}

// Get registers a route that matches 'GET' method.
func (s *Server) Get(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodGet, path, h, opts...)
}

// Head registers a route that matches 'HEAD' method.
func (s *Server) Head(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodHead, path, h, opts...)
}

// Options registers a route that matches 'OPTIONS' method.
func (s *Server) Options(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodOptions, path, h, opts...)
}

// Patch registers a route that matches 'PATCH' method.
func (s *Server) Patch(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodPatch, path, h, opts...)
}

// Post registers a route that matches 'POST' method.
func (s *Server) Post(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodPost, path, h, opts...)
}

// Put registers a route that matches 'PUT' method.
func (s *Server) Put(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodPut, path, h, opts...)
}

// Trace registers a route that matches 'TRACE' method.
func (s *Server) Trace(path string, h HandlerFunc, opts ...HandlerOption) {
	s.register(http.MethodTrace, path, h, opts...)
}

// Any registers a route that matches all the HTTP methods.
func (s *Server) Any(path string, handler HandlerFunc, opts ...HandlerOption) {
	s.Match(methods[:], path, handler, opts...)
}

// Match registers a route that matches specific methods.
func (s *Server) Match(methods []string, path string, handler HandlerFunc, opts ...HandlerOption) {
	info := newHandlerInfo(handler, opts)
	s.registerInfo(path, info, methods...)
}

// Handle registers routes from controller.
// It panics if controller's Kind is not struct.
func (s *Server) Handle(path string, controller interface{}, filters ...Filter) {
	//t := struct {
	//	Login  HandlerFunc `method:"get,post" path:"/login"`
	//	Logout HandlerFunc `path:"/logout"`
	//}{}
	v := reflect.Indirect(reflect.ValueOf(controller))
	if v.Kind() != reflect.Struct {
		panic("web > controller must be struct type")
	}

	t := v.Type()
	num := v.NumField()
	for i := 0; i < num; i++ {
		f := v.Field(i)
		h, ok := f.Interface().(HandlerFunc)
		if !ok {
			continue
		}

		sf := t.Field(i)
		if h == nil {
			panic(fmt.Sprintf("web > handler %s.%s isn't initialized", t.Name(), sf.Name))
		}

		s.handleField(path, t, &sf, h, filters...)
	}
}

func (s *Server) handleField(prefix string, t reflect.Type, sf *reflect.StructField, handler HandlerFunc, filters ...Filter) {
	var (
		p       string
		methods []string
		opts    = reflects.StructTag(sf.Tag).All()
		info    = &handlerInfo{
			action: handler.Chain(filters...),
		}
	)

	for k, v := range opts {
		switch k {
		case "name", "n":
			info.name = v
		case "path", "p":
			p = v
		case "authorize", "auth", "a":
			info.authorize = parseAuthorizeMode(v, s.cfg.Authorize)
		case "method", "m":
			methods = strings.Split(strings.ToUpper(v), ",")
		default:
			info.addOption(k, v)
		}
	}

	if info.name == "" {
		info.name = strings.ToLower(strings.TrimSuffix(t.Name(), "Controller")) + "." + texts.Rename(sf.Name, texts.Lower)
	}
	if p == "" {
		p = "/" + strings.ToLower(sf.Name)
	}

	if methods == nil {
		s.registerInfo(prefix+p, info, http.MethodGet)
	} else {
		s.registerInfo(prefix+p, info, methods...)
	}
}

// Static serves files from the given file system root.
func (s *Server) Static(prefix, root string) {
	//fs := http.StripPrefix(prefix, http.FileServer(http.Dir(root)))
	handler := func(c Context) error {
		return c.Content(path.Join(root, c.Path("")))
		//fs.ServeHTTP(c.Response(), c.Request())
		//return nil
	}
	p := path.Join(prefix, "/*")
	s.Head(p, handler)
	s.Get(p, handler)
}

// File registers a route in order to server a single file of the local filesystem.
func (s *Server) File(path, file string) {
	s.Get(path, func(c Context) error {
		return c.Content(file)
	})
}

// FileSystem serves files from a custom file system.
func (s *Server) FileSystem(prefix string, fs http.FileSystem) {
	fileServer := http.FileServer(fs)
	handler := func(c Context) error {
		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	p := path.Join(prefix, "/*")
	s.Head(p, handler)
	s.Get(p, handler)
}

func (s *Server) register(method, path string, handler HandlerFunc, opts ...HandlerOption) {
	info := newHandlerInfo(handler, opts)
	s.registerInfo(path, info, method)
}

func (s *Server) registerInfo(path string, info *handlerInfo, methods ...string) {
	for _, m := range methods {
		r, err := s.router.Add(m, path, info)
		if err != nil {
			panic(err)
		}

		if info.name == "" {
			info.name = m + ":" + path
		}
		if _, ok := s.routes[info.name]; ok {
			s.Logger.Warnf("web > A handler with name '%s' already exists", info.name)
		} else {
			s.routes[info.name] = r
		}
	}
}

// Group creates a new router group.
func (s *Server) Group(prefix string, filters ...Filter) (g *Group) {
	g = &Group{prefix: prefix, server: s}
	g.Use(filters...)
	return
}

// URL generates an URL from handler name and provided parameters.
func (s *Server) URL(name string, params ...interface{}) string {
	if r := s.routes[name]; r != nil {
		return r.URL(params)
	}
	return ""
}

func (s *Server) Handler(name string) (handler HandlerInfo) {
	if r := s.routes[name]; r != nil {
		handler = r.Handler().(*handlerInfo)
	}
	return
}

// AcquireContext returns an `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (s *Server) AcquireContext(w http.ResponseWriter, r *http.Request) Context {
	return s.ctxPool.Get(w, r)
}

// ReleaseContext returns the `Context` instance back to the pool.
func (s *Server) ReleaseContext(c Context) {
	s.ctxPool.Pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.ctxPool.Get(w, r)

	p := r.URL.EscapedPath()
	route, tsr := s.router.Find(r.Method, p, c.pathValues)
	if tsr && s.cfg.RedirectTrailingSlash {
		s.redirect(c, p)
	} else {
		s.execute(c, route)
	}

	s.ctxPool.Put(c)
}

func (s *Server) redirect(c *context, p string) {
	var code int
	if c.request.Method == http.MethodGet {
		code = http.StatusMovedPermanently
	} else {
		code = http.StatusTemporaryRedirect
	}

	if p[len(p)-1] == '/' {
		c.Status(code).Redirect(p[:len(p)-1])
	} else {
		c.Status(code).Redirect(p + "/")
	}
}

func (s *Server) execute(c *context, route router.Route) {
	if route != nil {
		if h := route.Handler(); h != nil {
			c.handler = h.(*handlerInfo)
		} else if s.cfg.MethodNotAllowed {
			c.handler = methodNotAllowed
		}
		c.pathNames = route.Params()
		c.route = route.Path()
	}

	// attach filters
	h := c.Handler().Action()
	for i := len(s.filters) - 1; i >= 0; i-- {
		h = s.filters[i].Apply(h)
	}

	// execute handler
	if err := h(c); err != nil {
		s.ErrorHandler.handle(c, err)
	}
}

// Serve starts the HTTP server.
func (s *Server) Serve() error {
	servers, err := s.buildServers()
	if err != nil {
		return err
	}

	s.servers = servers
	s.printRoutes()

	errs := make(chan error)
	for _, server := range servers {
		server.Handler = s
		go s.startServer(server, errs)
	}
	return <-errs
}

func (s *Server) buildServers() ([]*http.Server, error) {
	servers := make([]*http.Server, len(s.cfg.Entries))
	for i, entry := range s.cfg.Entries {
		tlsConfig, err := s.getTLSConfig(&entry)
		if err != nil {
			return nil, err
		}

		servers[i] = &http.Server{
			Addr:              entry.Address,
			ReadTimeout:       s.cfg.ReadTimeout,
			ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
			WriteTimeout:      s.cfg.WriteTimeout,
			IdleTimeout:       s.cfg.IdleTimeout,
			MaxHeaderBytes:    int(s.cfg.MaxHeaderSize),
			ErrorLog:          s.stdLogger,
			TLSConfig:         tlsConfig,
		}
	}
	return servers, nil
}

// StartServer runs a custom HTTP server.
func (s *Server) startServer(server *http.Server, errs chan error) {
	var (
		network string
		err     error
		ln      net.Listener
	)

	network, server.Addr = s.parseAddress(server.Addr)
	if network == "unix" {
		// TODO: handle error
		os.Remove(server.Addr)
	}
	ln, err = net.Listen(network, server.Addr)

	if err == nil {
		ln = tcpKeepAliveListener{ln.(*net.TCPListener)}
		if server.TLSConfig == nil {
			s.Logger.Infof("web > Server started on %s", server.Addr)
			err = server.Serve(ln)
		} else {
			s.Logger.Infof("web > Server started on %s(tls)", server.Addr)
			err = server.ServeTLS(ln, "", "")
		}
	}
	errs <- err
}

func (s *Server) parseAddress(addr string) (network, address string) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	if addr[0] == '/' {
		return "unix", addr
	}
	return "tcp", addr
}

func (s *Server) getTLSConfig(entry *Entry) (tlsConfig *tls.Config, err error) {
	if entry.TLS == nil {
		return
	}

	if entry.TLS.Cert != "" && entry.TLS.Key != "" {
		tlsConfig = new(tls.Config)
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(entry.TLS.Cert, entry.TLS.Key)
		if err != nil {
			return nil, errors.New("load TLS cert failed: " + err.Error())
		}
	} else if acme := entry.TLS.ACME; acme != nil {
		m := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
		}
		if domains := strings.Split(acme.Domain, ","); len(domains) > 0 {
			m.HostPolicy = autocert.HostWhitelist(domains...)
		}
		// todo: support default cache dir
		if acme.CacheDir == "" {
			s.Logger.Warn("ACME not using a cache: cache_dir is not set")
		} else {
			if err := os.MkdirAll(acme.CacheDir, 0700); err != nil {
				s.Logger.Warnf("ACME not using a cache: %v", err)
			} else {
				m.Cache = autocert.DirCache(acme.CacheDir)
			}
		}
		tlsConfig = &tls.Config{GetCertificate: m.GetCertificate}
	}
	return
}

func (s *Server) printRoutes() {
	s.router.Walk(func(r router.Route, m string) {
		handler := r.Handler().(*handlerInfo)
		s.Logger.Debugf("web > [%s] %s -> %s", texts.PadCenter(m, ' ', 7), r.Path(), handler.Name())
	})
}

// Shutdown gracefully shutdown the internal HTTP servers with timeout.
func (s *Server) Close(timeout time.Duration) {
	if timeout <= 0 {
		for _, server := range s.servers {
			if err := server.Close(); err != nil {
				s.Logger.Warnf("web > Server [%s] shutdown failed: %s", server.Addr, err)
			} else {
				s.Logger.Infof("web > Server [%s] shutdown successfully", server.Addr)
			}
		}
	} else {
		ctx, cancel := scontext.WithTimeout(scontext.Background(), timeout)
		defer cancel()

		var g sync.WaitGroup
		g.Add(len(s.servers))
		for _, server := range s.servers {
			go func(server *http.Server) {
				if err := server.Shutdown(ctx); err != nil {
					s.Logger.Warnf("web > Server [%s] shutdown failed: %s", server.Addr, err)
				} else {
					s.Logger.Infof("web > Server [%s] shutdown gracefully", server.Addr)
				}
				g.Done()
			}(server)
		}
		g.Wait()
	}
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

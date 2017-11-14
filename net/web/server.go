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
	"unsafe"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web/binder"
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
	Logger       *log.Logger
	stdLogger    *slog.Logger
	cfg          *Options
	acmeMgr      autocert.Manager
	filters      []Filter
	router       *router.Tree
	ctxPool      *contextPool
	servers      []*http.Server
}

// Default creates an instance of Server with default options.
func Default(address ...string) (server *Server) {
	c := &Options{Addresses: address}
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
		cfg:          c,
		Logger:       log.Get(PkgName),
		ErrorHandler: DefaultErrorHandler,
		Binder:       binder.New(binder.Options{MaxMemory: c.MaxBodySize}),
		router:       router.New(router.Options{}),
	}
	s.Logger.SetLevel(log.LevelDebug) // log.LevelOff
	s.stdLogger = slog.New(s.Logger, "web > ", 0)
	s.ctxPool = newContextPool(s)
	if c.ACME.Enabled {
		s.acmeMgr = autocert.Manager{
			Prompt: autocert.AcceptTOS,
		}
	}
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
func (s *Server) Connect(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodConnect, path, h, filters...)
}

// Delete registers a route that matches 'DELETE' method.
func (s *Server) Delete(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodDelete, path, h, filters...)
}

// Get registers a route that matches 'GET' method.
func (s *Server) Get(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodGet, path, h, filters...)
}

// Head registers a route that matches 'HEAD' method.
func (s *Server) Head(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodHead, path, h, filters...)
}

// Options registers a route that matches 'OPTIONS' method.
func (s *Server) Options(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodOptions, path, h, filters...)
}

// Patch registers a route that matches 'PATCH' method.
func (s *Server) Patch(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodPatch, path, h, filters...)
}

// Post registers a route that matches 'POST' method.
func (s *Server) Post(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodPost, path, h, filters...)
}

// Put registers a route that matches 'PUT' method.
func (s *Server) Put(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodPut, path, h, filters...)
}

// Trace registers a route that matches 'TRACE' method.
func (s *Server) Trace(path string, h HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.add(http.MethodTrace, path, h, filters...)
}

// Any registers a route that matches all the HTTP methods.
func (s *Server) Any(path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	return s.Match(methods[:], path, handler, filters...)
}

// Match registers a route that matches specific methods.
func (s *Server) Match(methods []string, path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	info := &handlerInfo{
		name:   handler.Name(),
		action: handler.Chain(filters...),
	}
	s.addInfo(path, info, methods...)
	return info
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

		st := reflects.StructTag(sf.Tag)
		a := parseAuthorizeMode(st.Find("authorize", "a"), s.cfg.Authorize)
		n := st.Find("name", "n")
		p := st.Find("path", "p")
		if n == "" {
			n = strings.ToLower(t.PkgPath() + "." + strings.TrimSuffix(t.Name(), "Controller") + "." + sf.Name)
		}
		if p == "" {
			p = "/" + strings.ToLower(sf.Name)
		}
		p = path + p

		if method := st.Find("method", "m"); method != "" {
			methods := strings.Split(strings.ToUpper(method), ",")
			s.Match(methods, p, h, filters...).SetName(n).SetAuthorize(a)
		} else {
			s.Get(p, h, filters...).SetName(n).SetAuthorize(a)
		}
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

func (s *Server) add(method, path string, handler HandlerFunc, filters ...Filter) HandlerCustomizer {
	info := &handlerInfo{
		name:   handler.Name(),
		action: handler.Chain(filters...),
	}
	s.addInfo(path, info, method)
	return info
}

func (s *Server) addInfo(path string, info *handlerInfo, methods ...string) {
	err := s.router.Add(path, unsafe.Pointer(info), methods...)
	if err != nil {
		panic(err)
	}
}

// Group creates a new router group.
func (s *Server) Group(prefix string, filters ...Filter) (g *Group) {
	g = &Group{prefix: prefix, server: s}
	g.Use(filters...)
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
		if ptr := route.Handler(); ptr != nil {
			c.handler = (*handlerInfo)(ptr)
		} else {
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
		s.ErrorHandler.Handle(c, err)
	}
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	if len(s.cfg.Addresses) == 0 {
		return errors.New("web: listen address is empty")
	}

	servers, err := s.buildServers()
	if err != nil {
		return err
	}

	s.servers = servers
	s.printRoutes()

	errs := make(chan error)
	for _, server := range servers {
		go s.startServer(server, errs)
	}
	return <-errs
}

func (s *Server) buildServers() ([]*http.Server, error) {
	tlsConfig, err := s.getTLSConfig()
	if err != nil {
		return nil, err
	}

	servers := make([]*http.Server, len(s.cfg.Addresses))
	for i, addr := range s.cfg.Addresses {
		servers[i] = &http.Server{
			Addr:              addr,
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
	server.Handler = s

	var (
		addrType string
		err      error
		l        net.Listener
	)

	addrType, server.Addr = s.parseAddress(server.Addr)
	switch addrType {
	case "https":
		s.Logger.Infof("web > https server started on %s", server.Addr)
		err = server.ListenAndServeTLS("", "")
	case "unix":
		os.Remove(server.Addr)
		if l, err = net.Listen("unix", server.Addr); err == nil {
			s.Logger.Infof("web > unix server started on %s", server.Addr)
			err = server.Serve(l)
		}
	default:
		s.Logger.Infof("web > http server started on %s", server.Addr)
		err = server.ListenAndServe()
	}
	errs <- err
}

func (s *Server) parseAddress(addr string) (addrType, addrValue string) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "http", addr
}

func (s *Server) printRoutes() {
	s.router.Walk(func(r router.Route, m string) {
		handler := (*handlerInfo)(r.Handler())
		s.Logger.Infof("web > [%s] %s -> %s", texts.PadCenter(m, ' ', 7), r.Path(), handler.Name())
	})
}

func (s *Server) getTLSConfig() (tlsConfig *tls.Config, err error) {
	c := s.cfg
	if c.TLSCertFile != "" && c.TLSKeyFile != "" {
		tlsConfig = new(tls.Config)
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
		if err != nil {
			return nil, errors.New("load tls cert failed: " + err.Error())
		}
	} else if c.ACME.Enabled {
		tlsConfig = new(tls.Config)
		tlsConfig.GetCertificate = s.acmeMgr.GetCertificate
	}
	if tlsConfig != nil && !s.cfg.TLSDisableHTTP2 {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
	}
	return
}

// Shutdown gracefully shutdown the HTTP server with timeout.
func (s *Server) Close(timeout time.Duration) {
	if timeout <= 0 {
		for _, server := range s.servers {
			if err := server.Close(); err != nil {
				s.Logger.Warnf("web > Server [%s] shutdown failed: %s", server.Addr, err)
			} else {
				s.Logger.Infof("web > Server [%s] shutdown initiated", server.Addr)
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
					s.Logger.Infof("web > Server [%s] shutdown initiated", server.Addr)
				}
				g.Done()
			}(server)
		}
		g.Wait()
	}
}

package web

import (
	ctx "context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/util/cast"
)

// Requester holds all methods for reading request.
type Requester interface {
	// Request returns `*http.Request`.
	Request() *http.Request

	// SetRequest replace default `*http.Request`.
	SetRequest(r *http.Request)

	// IsTLS returns true if HTTP connection is TLS otherwise false.
	IsTLS() bool

	// Scheme returns the HTTP protocol scheme, `http` or `https`.
	Scheme() string

	// RealIP returns the client's network address based on `X-Forwarded-For`
	// or `X-Real-IP` request header.
	RealIP() string

	// IsAJAX returns true if this request is an AJAX request(XMLHttpRequest)
	IsAJAX() bool

	// Path returns path parameter by name.
	Path(name string) string

	// P returns path parameter by name, it's an alias of Path method.
	P(name string) string

	// ParamNames returns path parameter names.
	PathNames() []string

	// SetPathNames sets path parameter names.
	SetPathNames(names ...string)

	// PathValues returns path parameter values.
	PathValues() []string

	// SetPathValues sets path parameter values.
	SetPathValues(values ...string)

	// Query returns the query param for the provided name.
	Query(name string) string

	// Q returns the query param for the provided name, it's an alias of Query method.
	Q(name string) string

	// QueryValues returns the query parameters as `url.Values`.
	QueryValues() url.Values

	// Form returns the form field value for the provided name.
	Form(name string) string

	// F returns the form field value for the provided name, it's an alias of Form method.
	F(name string) string

	// FormValues returns the form parameters as `url.Values`.
	FormValues() (url.Values, error)

	// File returns the multipart form file for the provided name.
	File(name string) (multipart.File, *multipart.FileHeader, error)

	// MultipartForm returns the multipart form.
	MultipartForm() (*multipart.Form, error)

	// Cookie returns the named cookie provided in the request.
	Cookie(name string) (*http.Cookie, error)

	// Header returns request's header by name.
	Header(name string) string

	// ContentType returns the 'Content-Type' header of request. Charset is omited.
	ContentType() string
}

// Responser holds all methods for sending response.
type Responser interface {
	// Response returns `ResponseWriter`.
	Response() ResponseWriter

	// Status set status code of response.
	Status(code int) Responser

	// SetCookie adds a `Set-Cookie` header in HTTP response.
	SetCookie(cookie *http.Cookie) Responser

	// SetHeader writes given key and value to the response's header.
	SetHeader(key, value string) Responser

	// SetContentType sets the 'Content-Type' header of response.
	SetContentType(ct string) Responser

	// Render renders a template with data and sends a text/html response.
	// Renderer must be registered using `Server.Renderer`.
	Render(name string, data interface{}) error

	// HTML sends an text/html response.
	HTML(html string) error

	// JSON sends a JSON response.
	JSON(i interface{}, indent ...string) error

	// JSONP sends a JSONP response. It uses `callback` to construct the JSONP payload.
	JSONP(callback string, i interface{}) error

	// XML sends an XML response.
	XML(i interface{}, indent ...string) error

	// Text sends string as plain text.
	Text(s string) error

	// Data sends byte array to response.
	// You can use `Context.SetContentType` to set content type.
	Data(b []byte, cd ...ContentDisposition) error

	// Stream sends a streaming response.
	// You can use `Context.SetContentType` to set content type.
	Stream(r io.Reader, cd ...ContentDisposition) error

	// Content sends a response with the content of the file.
	Content(file string, cd ...ContentDisposition) error

	// Empty sends a response with no body.
	// You can use `Context.Status` to change status code (default is 200).
	Empty() error

	// Redirect redirects the request to a provided URL with status code.
	// You can use `Context.Status` to change status code.
	// If code wasn't set, default is 302 (http.StatusFound).
	Redirect(url string) error
}

// Context represents the context of the current HTTP request.
type Context interface {
	// Reset resets the context after request completes. It must be called along
	// with `Server#AcquireContext()` and `Server#ReleaseContext()`.
	// See `Server#ServeHTTP()`
	Reset(r *http.Request, w http.ResponseWriter)

	// Server returns the `Server` instance.
	Server() *Server

	// Logger returns the `Logger` instance.
	Logger() log.Logger

	// User returns info of current visitor.
	User() User

	// User set user info of current visitor. Generally used by authentication filter.
	SetUser(user User)

	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, value interface{})

	// Bind binds the request body into `i`.
	// Validator must be registered using `Server#Validator` if validate arg is true.
	Bind(i interface{}, validate ...bool) error

	// Route returns the registered route path for current handler.
	Route() string

	// Handler returns the matched handler info by router.
	Handler() HandlerInfo

	// Error invokes the registered HTTP error handler. Generally used by filters.
	Error(err error)

	Requester

	Responser

	ctx.Context
}

type context struct {
	request    *http.Request
	response   *responseWriter
	route      string
	pathNames  []string
	pathValues []string
	query      url.Values
	handler    HandlerInfo
	user       User
	data       data.Map
	server     *Server
}

func (c *context) Deadline() (deadline time.Time, ok bool) {
	return c.request.Context().Deadline()
}

func (c *context) Done() <-chan struct{} {
	return c.request.Context().Done()
}

func (c *context) Err() error {
	return c.request.Context().Err()
}

func (c *context) Value(key interface{}) interface{} {
	return c.request.Context().Value(key)
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) Response() ResponseWriter {
	return c.response
}

func (c *context) IsAJAX() bool {
	return c.request.Header.Get(HeaderXRequestedWith) == "XMLHttpRequest"
}

func (c *context) Header(name string) string {
	return c.request.Header.Get(name)
}

func (c *context) ContentType() string {
	ct := c.request.Header.Get(HeaderContentType)
	if ct != "" {
		if i := strings.Index(ct, ";"); i > 0 {
			ct = ct[:i]
		}
	}
	return ct
}

func (c *context) IsTLS() bool {
	return c.request.TLS != nil
}

func (c *context) Scheme() string {
	if c.IsTLS() {
		return "https"
	}
	return "http"
}

func (c *context) RealIP() string {
	ra := c.request.RemoteAddr
	if ip := strings.TrimSpace(c.request.Header.Get(HeaderXRealIP)); ip != "" {
		ra = ip
	} else if ip := strings.TrimSpace(c.request.Header.Get(HeaderXForwardedFor)); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

func (c *context) Route() string {
	return c.route
}

func (c *context) Path(name string) string {
	for i, n := range c.pathNames {
		if i < len(c.pathValues) {
			if n == name {
				return c.pathValues[i]
			}
		}
	}
	return ""
}

func (c *context) P(name string) string {
	return c.Path(name)
}

func (c *context) PathNames() []string {
	return c.pathNames
}

func (c *context) SetPathNames(names ...string) {
	c.pathNames = names
}

func (c *context) PathValues() []string {
	return c.pathValues
}

func (c *context) SetPathValues(values ...string) {
	c.pathValues = values
}

func (c *context) Query(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *context) Q(name string) string {
	return c.Query(name)
}

func (c *context) QueryValues() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) Form(name string) string {
	return c.request.FormValue(name)
}

func (c *context) F(name string) string {
	return c.request.FormValue(name)
}

func (c *context) FormValues() (url.Values, error) {
	if strings.HasPrefix(c.ContentType(), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(int64(c.server.cfg.MaxBodySize)); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

func (c *context) File(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.request.FormFile(name)
}

func (c *context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(int64(c.server.cfg.MaxBodySize))
	return c.request.MultipartForm, err
}

func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *context) SetCookie(cookie *http.Cookie) Responser {
	http.SetCookie(c.Response(), cookie)
	return c
}

func (c *context) User() User {
	return c.user
}

func (c *context) SetUser(user User) {
	c.user = user
}

func (c *context) Get(key string) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

func (c *context) Set(key string, val interface{}) {
	if c.data == nil {
		c.data = make(data.Map)
	}
	c.data[key] = val
}

func (c *context) Bind(i interface{}, validate ...bool) (err error) {
	if c.server.Binder == nil {
		return ErrBinderNotRegistered
	}

	err = c.server.Binder.Bind(c, i)
	if err == nil && len(validate) > 0 && validate[0] {
		if c.server.Validator == nil {
			return ErrValidatorNotRegistered
		}
		err = c.server.Validator.Validate(i)
	}
	return
}

func (c *context) Render(name string, data interface{}) (err error) {
	if c.server.Renderer == nil {
		return ErrRendererNotRegistered
	}
	c.response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	return c.server.Renderer.Render(c.response, name, data, c)
	// TODO: use buffer pool
	//buf := new(bytes.Buffer)
	//if err = c.server.Renderer.Render(buf, name, data, c); err != nil {
	//	return
	//}
	//return c.Data(MIMETextHTMLCharsetUTF8, buf.Bytes())
}

func (c *context) HTML(html string) (err error) {
	b := cast.StringToBytes(html)
	return c.SetContentType(MIMETextHTMLCharsetUTF8).Data(b)
}

func (c *context) Text(text string) (err error) {
	b := cast.StringToBytes(text)
	return c.SetContentType(MIMETextPlainCharsetUTF8).Data(b)
}

func (c *context) JSON(i interface{}, indent ...string) (err error) {
	var b []byte
	if len(indent) > 0 {
		b, err = json.MarshalIndent(i, "", indent[0])
	} else {
		b, err = json.Marshal(i)
	}

	if err == nil {
		return c.JSONBlob(b)
	}
	return
}

func (c *context) JSONBlob(b []byte) (err error) {
	return c.SetContentType(MIMEApplicationJSONCharsetUTF8).Data(b)
}

func (c *context) JSONP(callback string, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	c.response.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return err
	}
	if _, err = c.response.Write(b); err != nil {
		return err
	}
	_, err = c.response.Write([]byte(");"))
	return err
}

func (c *context) XML(i interface{}, indent ...string) (err error) {
	var b []byte
	if len(indent) > 0 {
		b, err = xml.MarshalIndent(i, "", indent[0])
	} else {
		b, err = xml.Marshal(i)
	}

	if err == nil {
		return c.XMLBlob(b)
	}
	return
}

func (c *context) XMLBlob(b []byte) (err error) {
	c.SetContentType(MIMEApplicationXMLCharsetUTF8)
	if _, err = c.response.Write([]byte(xml.Header)); err == nil {
		_, err = c.response.Write(b)
	}
	return
}

func (c *context) Data(b []byte, cd ...ContentDisposition) (err error) {
	if len(cd) > 0 {
		c.SetHeader(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", cd[0].Type, cd[0].Name))
	}
	_, err = c.response.Write(b)
	return
}

func (c *context) Stream(r io.Reader, cd ...ContentDisposition) (err error) {
	if len(cd) > 0 {
		c.SetHeader(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", cd[0].Type, cd[0].Name))
	}
	_, err = io.Copy(c.response, r)
	return
}

func (c *context) Content(file string, cd ...ContentDisposition) error {
	if len(cd) > 0 {
		c.SetHeader(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", cd[0].Type, cd[0].Name))
	}

	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		var p string
		for _, page := range c.server.cfg.IndexPages {
			file = filepath.Join(file, page)
			if fi, err = os.Stat(file); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return err
			}

			p = file
			break
		}

		if p == "" {
			return ErrNotFound
		}

		if f, err = os.Open(p); err != nil {
			return err
		}
		defer f.Close()
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return nil
}

func (c *context) Empty() error {
	c.response.CommitHeader()
	return nil
}

func (c *context) Redirect(url string) (err error) {
	status := c.response.status
	if status == http.StatusOK {
		status = http.StatusFound
	} else if status < http.StatusMultipleChoices || status > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	http.Redirect(c.response, c.request, url, status)
	return
}

func (c *context) SetHeader(key, value string) Responser {
	c.response.Header().Set(key, value)
	return c
}

func (c *context) SetContentType(value string) Responser {
	c.SetHeader(HeaderContentType, value)
	return c
}

func (c *context) Status(code int) Responser {
	c.response.status = code
	//c.response.WriteHeader(code)
	return c
}

func (c *context) Error(err error) {
	c.server.ErrorHandler.handle(c, err)
}

func (c *context) Server() *Server {
	return c.server
}

func (c *context) Handler() HandlerInfo {
	return c.handler
}

func (c *context) Logger() log.Logger {
	return c.server.Logger
}

func (c *context) Reset(r *http.Request, w http.ResponseWriter) {
	c.query = nil
	c.data = nil
	c.user = nil
	c.request = r
	c.response.reset(w)
	c.handler = notFound
}

type contextPool struct {
	sync.Pool
}

func newContextPool(s *Server) *contextPool {
	p := &contextPool{}
	p.New = func() interface{} {
		return &context{
			server:     s,
			response:   &responseWriter{server: s},
			pathValues: make([]string, s.router.MaxParam()),
		}
	}
	return p
}

func (p *contextPool) Get(w http.ResponseWriter, r *http.Request) (c *context) {
	c = p.Pool.Get().(*context)
	c.Reset(r, w)
	return
}

func (p *contextPool) Put(c *context) {
	p.Pool.Put(c)
}

func (p *contextPool) Call(w http.ResponseWriter, r *http.Request, fn func(c *context)) {
	c := p.Get(w, r)
	fn(c)
	p.Put(c)
}

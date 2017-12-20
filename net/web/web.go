// Package web implement a lightweight HTTP server.
package web

import (
	"io"
	"net/http"

	"github.com/cuigh/auxo/security"
)

const PkgName = "auxo.net.web"

const (
	charsetUTF8 = "charset=UTF-8"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

// Headers
const (
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAcceptLanguage      = "Accept-Language"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"
	HeaderExpires             = "Expires"
	HeaderCacheControl        = "Cache-Control"

	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

const (
	DispositionAttachment = "attachment"
	DispositionInline     = "inline"
)

var (
	methods = [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
)

type ContentDisposition struct {
	Type string
	Name string
}

type User = security.User

// Binder is the interface that can unmarshal request data to struct.
type Binder interface {
	// Bind takes data out of the request and decodes into a struct according
	// to the Content-Type of the request.
	Bind(ctx Context, i interface{}) (err error)
}

// Validator is the interface that check data is valid or not.
type Validator interface {
	Validate(i interface{}) error
}

// Renderer is the interface that render HTML template.
type Renderer interface {
	Render(io.Writer, string, interface{}, Context) error
}

type Router interface {
	Group(prefix string, filters ...Filter) *Group
	Use(filters ...Filter)

	Get(path string, h HandlerFunc, opts ...HandlerOption)
	Post(path string, h HandlerFunc, opts ...HandlerOption)
	Delete(path string, h HandlerFunc, opts ...HandlerOption)
	Put(path string, h HandlerFunc, opts ...HandlerOption)
	Head(path string, h HandlerFunc, opts ...HandlerOption)
	Options(path string, h HandlerFunc, opts ...HandlerOption)
	Patch(path string, h HandlerFunc, opts ...HandlerOption)
	Trace(path string, h HandlerFunc, opts ...HandlerOption)
	Connect(path string, h HandlerFunc, opts ...HandlerOption)
	Any(path string, handler HandlerFunc, opts ...HandlerOption)
	Match(methods []string, path string, h HandlerFunc, opts ...HandlerOption)
	Handle(path string, controller interface{}, filters ...Filter)

	Static(prefix, root string)
	File(path, file string)
	FileSystem(path string, fs http.FileSystem)
}

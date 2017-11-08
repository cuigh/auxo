package web

import (
	"net/http"
	"reflect"
	"runtime"

	"github.com/cuigh/auxo/errors"
)

var (
	notFound         = &handlerInfo{action: WrapError(ErrNotFound)}
	methodNotAllowed = &handlerInfo{action: WrapError(ErrMethodNotAllowed)}
)

// WrapError wraps `http.Handler` into `web.HandlerFunc`.
func WrapError(err *Error) HandlerFunc {
	return func(c Context) error {
		return err
	}
}

// WrapHandler wraps `http.Handler` into `web.HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// Filter defines a filter interface.
type Filter interface {
	Apply(HandlerFunc) HandlerFunc
}

// WrapFilter wraps `func(http.Handler) http.Handler` into `web.FilterFunc`
func WrapFilter(f func(http.Handler) http.Handler) Filter {
	fn := func(next HandlerFunc) HandlerFunc {
		return func(c Context) (err error) {
			f(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.(*context).request = r
				err = next(c)
			})).ServeHTTP(c.Response(), c.Request())
			return
		}
	}
	return FilterFunc(fn)
}

// FilterFunc defines a filter function type.
type FilterFunc func(HandlerFunc) HandlerFunc

func (f FilterFunc) Apply(h HandlerFunc) HandlerFunc {
	return f(h)
}

// HandlerFunc defines a function to server HTTP requests.
type HandlerFunc func(Context) error

func (h HandlerFunc) Name() string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

func (h HandlerFunc) Chain(filters ...Filter) HandlerFunc {
	handler := h
	for i := len(filters) - 1; i >= 0; i-- {
		handler = filters[i].Apply(handler)
	}
	return handler
}

// HandlerCustomizer is the interface that customizes handler.
type HandlerCustomizer interface {
	SetName(name string) HandlerCustomizer
	SetAuthorize(mode AuthorizeMode) HandlerCustomizer
	SetOption(name, value string) HandlerCustomizer
}

// HandlerInfo is the interface for handler info.
type HandlerInfo interface {
	Action() HandlerFunc
	Name() string
	Authorize() AuthorizeMode
	Option(name string) string
}

const (
	AuthAnonymous     AuthorizeMode = iota // everyone
	AuthAuthenticated                      // all logged-in user
	AuthExplicit                           // must be explicit granted
)

type AuthorizeMode uint8

func parseAuthorizeMode(s string, def AuthorizeMode) AuthorizeMode {
	switch s {
	case "?", "authenticated":
		return AuthAuthenticated
	case "!", "explicit":
		return AuthExplicit
	case "*", "anonymous":
		return AuthAnonymous
	default:
		return def
	}
}

func (a *AuthorizeMode) Unmarshal(i interface{}) error {
	if s, ok := i.(string); ok {
		*a = parseAuthorizeMode(s, AuthAuthenticated)
		return nil
	}
	return errors.Format("can't convert %v to authorizeMode", i)
}

func (a AuthorizeMode) String() string {
	switch a {
	case AuthAuthenticated:
		return "authenticated"
	case AuthExplicit:
		return "explicit"
	default:
		return "anonymous"
	}
}

type handlerInfo struct {
	action    HandlerFunc
	name      string
	authorize AuthorizeMode
	options   map[string]string
}

func (h *handlerInfo) Action() HandlerFunc {
	return h.action
}

func (h *handlerInfo) Name() string {
	return h.name
}

func (h *handlerInfo) Authorize() AuthorizeMode {
	return h.authorize
}

func (h *handlerInfo) SetName(name string) HandlerCustomizer {
	h.name = name
	return h
}

func (h *handlerInfo) SetAuthorize(mode AuthorizeMode) HandlerCustomizer {
	h.authorize = mode
	return h
}

func (h *handlerInfo) Option(name string) string {
	if h.options == nil {
		return ""
	}
	return h.options[name]
}

func (h *handlerInfo) SetOption(name, value string) HandlerCustomizer {
	if h.options == nil {
		h.options = make(map[string]string)
	}
	h.options[name] = value
	return h
}

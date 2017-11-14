package filter

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cuigh/auxo/net/web"
)

// CORS represents a Cross-Origin Resource Sharing (CORS) filter.
type CORS struct {
	// AllowOrigin defines a list of origins that may access the resource.
	// Optional. Default value []string{"*"}.
	AllowOrigins []string

	// AllowMethods defines a list methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	// Optional. Default value DefaultCORSConfig.AllowMethods.
	AllowMethods []string

	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This in response to a preflight request.
	// Optional. Default value []string{}.
	AllowHeaders []string

	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	// Optional. Default value false.
	AllowCredentials bool

	// ExposeHeaders defines a whitelist headers that clients are allowed to
	// access.
	// Optional. Default value []string{}.
	ExposeHeaders []string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	// Optional. Default value 0.
	MaxAge int
}

// Apply implements `web.Filter` interface.
func (c *CORS) Apply(next web.HandlerFunc) web.HandlerFunc {
	if len(c.AllowOrigins) == 0 {
		c.AllowOrigins = []string{"*"}
	}
	if len(c.AllowMethods) == 0 {
		c.AllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete}
	}

	allowMethods := strings.Join(c.AllowMethods, ",")
	allowHeaders := strings.Join(c.AllowHeaders, ",")
	exposeHeaders := strings.Join(c.ExposeHeaders, ",")
	maxAge := strconv.Itoa(c.MaxAge)

	return func(ctx web.Context) error {
		r := ctx.Request()
		w := ctx.Response()
		origin := r.Header.Get(web.HeaderOrigin)
		if origin == "" {
			return next(ctx)
		}

		// Check allowed origins
		var allowOrigin string
		for _, o := range c.AllowOrigins {
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
		}

		// Simple request
		if r.Method != http.MethodOptions {
			w.Header().Add(web.HeaderVary, web.HeaderOrigin)
			w.Header().Set(web.HeaderAccessControlAllowOrigin, allowOrigin)
			if c.AllowCredentials {
				w.Header().Set(web.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				w.Header().Set(web.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			return next(ctx)
		}

		// Preflight request
		w.Header().Add(web.HeaderVary, web.HeaderOrigin)
		w.Header().Add(web.HeaderVary, web.HeaderAccessControlRequestMethod)
		w.Header().Add(web.HeaderVary, web.HeaderAccessControlRequestHeaders)
		w.Header().Set(web.HeaderAccessControlAllowOrigin, allowOrigin)
		w.Header().Set(web.HeaderAccessControlAllowMethods, allowMethods)
		if c.AllowCredentials {
			w.Header().Set(web.HeaderAccessControlAllowCredentials, "true")
		}
		if allowHeaders != "" {
			w.Header().Set(web.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := r.Header.Get(web.HeaderAccessControlRequestHeaders)
			if h != "" {
				w.Header().Set(web.HeaderAccessControlAllowHeaders, h)
			}
		}
		if c.MaxAge > 0 {
			w.Header().Set(web.HeaderAccessControlMaxAge, maxAge)
		}
		return ctx.Status(http.StatusNoContent).Empty()
	}
}

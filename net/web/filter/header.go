package filter

import (
	"strconv"
	"time"

	"github.com/cuigh/auxo/net/web"
)

// Header is a filter which can inject headers to response.
type Header map[string]string

func (h Header) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx web.Context) error {
		for key, value := range h {
			ctx.SetHeader(key, value)
		}
		return next(ctx)
	}
}

// DynamicHeader is a filter which can inject dynamic headers to response.
type DynamicHeader map[string]func() string

// NewCacheHeader return a Header instance which adds `Expires` and `Cache-Control` headers.
func NewCacheHeader(expiry time.Duration) DynamicHeader {
	maxAge := "max-age=" + strconv.Itoa(int(expiry.Seconds()))
	return DynamicHeader{
		web.HeaderExpires: func() string {
			return time.Now().Add(expiry).Format(time.RFC1123)
		},
		web.HeaderCacheControl: func() string { return maxAge },
	}
}

func (h DynamicHeader) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx web.Context) error {
		for key, fn := range h {
			ctx.SetHeader(key, fn())
		}
		return next(ctx)
	}
}

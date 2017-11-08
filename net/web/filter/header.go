package filter

import (
	"strconv"
	"time"

	"github.com/cuigh/auxo/net/web"
)

// Header is a filter which can inject headers to response.
type Header map[string]string

// NewCacheHeader return a Header instance which adds `Expires` and `Cache-Control` headers.
func NewCacheHeader(expiry time.Duration) Header {
	return Header{
		web.HeaderExpires:      time.Now().Add(expiry).Format(time.RFC1123),
		web.HeaderCacheControl: "max-age=" + strconv.Itoa(int(expiry.Seconds())),
	}
}

func (h Header) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx web.Context) error {
		for key, value := range h {
			ctx.SetHeader(key, value)
		}
		return next(ctx)
	}
}

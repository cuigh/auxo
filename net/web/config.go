package web

import (
	"time"

	"github.com/cuigh/auxo"
	"github.com/cuigh/auxo/byte/size"
)

type Options struct {
	Debug           bool
	Name            string
	Mode            string   // dev/prd
	Addresses       []string `option:"address"` // http://:8001,https://auxo.net:443,unix:///a/b
	TLSCertFile     string
	TLSKeyFile      string
	TLSDisableHTTP2 bool
	ACME            struct { // default: Let's Encrypt
		Enabled bool
		Host    string
		Dir     string
	}
	RedirectTrailingSlash bool
	ReadTimeout           time.Duration
	ReadHeaderTimeout     time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	MaxHeaderSize         size.Size
	MaxBodySize           size.Size
	Authorize             AuthorizeMode // default authorize mode
	IndexUrl              string
	LoginUrl              string
	UnauthorizedUrl       string
	IndexPages            []string // static index pages, default: index.html
}

func (opts *Options) ensure() {
	if len(opts.Addresses) == 0 {
		opts.Addresses = []string{"http://:8080"}
	}
	if opts.MaxBodySize <= 0 {
		opts.MaxBodySize = 10 * size.MB
	}
	if opts.Name == "" {
		opts.Name = "auxo/" + auxo.Version
	}
	if len(opts.IndexPages) == 0 {
		opts.IndexPages = []string{"index.html"}
	}
}

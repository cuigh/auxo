package web

import (
	"time"

	"github.com/cuigh/auxo"
	"github.com/cuigh/auxo/byte/size"
)

// Entry represents an entry-point for listening.
type Entry struct {
	Address string // http://:8001,https://auxo.net:443,unix:///a/b
	TLS     *struct {
		Cert string
		Key  string
		ACME *struct {
			//Server string // default: https://acme-staging.api.letsencrypt.org/directory
			Domain   string
			Email    string
			CacheDir string
		}
	}
}

// Entry represents the options of Server.
type Options struct {
	// Name is used as HTTP `Server` header, default is auxo/[auxo.Version]
	Name                  string
	Mode                  string // develop/release
	Entries               []Entry
	RedirectTrailingSlash bool
	MethodNotAllowed      bool
	ReadTimeout           time.Duration
	ReadHeaderTimeout     time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	MaxHeaderSize         size.Size
	MaxBodySize           size.Size
	Authorize             AuthorizeMode // default authorize mode
	IndexPages            []string      // static index pages, default: index.html
	//IndexUrl              string
	//LoginUrl              string
	//UnauthorizedUrl       string
}

func (opts *Options) ensure() {
	if len(opts.Entries) == 0 {
		opts.Entries = []Entry{{Address: ":80"}}
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

package jet

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/net/web"
	"github.com/cuigh/auxo/net/web/renderer"
	"io"
	"io/fs"
	"path/filepath"
)

type Renderer struct {
	set *jet.Set
}

type fsLoader struct {
	fs fs.FS
}

func (l fsLoader) Open(name string) (io.ReadCloser, error) {
	// name start with '/'
	return l.fs.Open(name[1:])
}

func (l fsLoader) Exists(name string) bool {
	// name start with '/'
	fi, err := fs.Stat(l.fs, name[1:])
	return err == nil && !fi.IsDir()
}

type Options struct {
	debug bool
	fs    fs.FS
	dir   string
	vars  map[string]interface{}
}

func (opts *Options) ensure() (err error) {
	if opts.fs == nil && opts.dir == "" {
		d := filepath.Dir(app.Path())
		p := filepath.Join(d, "views")
		if files.Exist(p) {
			opts.dir = p
		} else if p = filepath.Join(d, "resources/views"); files.Exist(p) {
			opts.dir = p
		} else {
			return errors.New("std: can't locate templates directory")
		}
	}
	return nil
}

type Option func(opts *Options)

func Debug(b ...bool) Option {
	return func(opts *Options) {
		opts.debug = len(b) == 0 || b[0]
	}
}

func Dir(f fs.FS, dir string) Option {
	return func(opts *Options) {
		opts.fs = f
		opts.dir = dir
	}
}

func Var(name string, value interface{}) Option {
	return func(opts *Options) {
		opts.vars[name] = value
	}
}

func VarMap(m map[string]interface{}) Option {
	return func(opts *Options) {
		for k, v := range m {
			opts.vars[k] = v
		}
	}
}

func New(opts ...Option) (*Renderer, error) {
	options := Options{
		vars: make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(&options)
	}
	err := options.ensure()
	if err != nil {
		return nil, err
	}

	var loader jet.Loader
	if options.fs == nil {
		loader = jet.NewOSFileSystemLoader(options.dir)
	} else {
		if options.dir != "" {
			options.fs, err = fs.Sub(options.fs, options.dir)
			if err != nil {
				return nil, err
			}
		}
		loader = fsLoader{fs: options.fs}
	}

	var set *jet.Set
	if options.debug {
		set = jet.NewSet(loader, jet.InDevelopmentMode())
	} else {
		set = jet.NewSet(loader)
	}
	// Add common functions
	set.AddGlobalFunc("eq", equal)
	set.AddGlobalFunc("choose", choose)
	set.AddGlobal("printf", fmt.Sprintf)
	set.AddGlobal("limit", renderer.Limit)
	set.AddGlobal("slice", renderer.Slice)
	for k, v := range options.vars {
		set.AddGlobal(k, v)
	}
	return &Renderer{set: set}, nil
}

func Must(opts ...Option) *Renderer {
	r, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, ctx web.Context) error {
	tpl, err := r.set.GetTemplate(name)
	if err == nil {
		err = tpl.Execute(w, nil, data)
	}
	return err
}

package jet

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/CloudyKit/jet"
	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/net/web"
	"github.com/cuigh/auxo/net/web/renderer"
)

type Renderer struct {
	set *jet.Set
}

type Options struct {
	debug bool
	dirs  []string
	vars  map[string]interface{}
}

func (opts *Options) ensure() error {
	if len(opts.dirs) == 0 {
		d := filepath.Dir(app.Path())
		p := filepath.Join(d, "views")
		if files.Exist(p) {
			opts.dirs = append(opts.dirs, p)
		} else {
			p = filepath.Join(d, "resources/views")
			if files.Exist(p) {
				opts.dirs = append(opts.dirs, p)
			}
		}
	}
	if len(opts.dirs) == 0 {
		return errors.New("jet: can't locate templates directory")
	}
	return nil
}

type Option func(opts *Options)

func Dir(dir ...string) Option {
	return func(opts *Options) {
		opts.dirs = append(opts.dirs, dir...)
	}
}

func Debug(b ...bool) Option {
	return func(opts *Options) {
		opts.debug = len(b) == 0 || b[0]
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

	set := jet.NewHTMLSet(options.dirs...)
	set.SetDevelopmentMode(options.debug)
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

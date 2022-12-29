package std

import (
	"errors"
	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/net/web"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Options struct {
	debug bool
	fs    fs.FS
	dir   string
	exts  []string
	fm    template.FuncMap
}

func (opts *Options) ensure() error {
	if opts.fs == nil {
		if opts.dir == "" {
			d := filepath.Dir(app.Path())
			p := filepath.Join(d, "views")
			if files.Exist(p) {
				opts.dir = p
			} else {
				p = filepath.Join(d, "resources/views")
				if files.Exist(p) {
					opts.dir = p
				}
			}
		}
		if opts.dir == "" {
			return errors.New("std: can't locate templates directory")
		}
	} else {
		if opts.dir == "" {
			opts.dir = "."
		}
	}

	if len(opts.exts) == 0 {
		opts.exts = []string{".html", ".gohtml"}
	}
	return nil
}

type Option func(opts *Options)

func Dir(fs fs.FS, dir string) Option {
	return func(opts *Options) {
		opts.fs = fs
		opts.dir = dir
	}
}

func Ext(ext ...string) Option {
	return func(opts *Options) {
		opts.exts = ext
	}
}

func Debug(b ...bool) Option {
	return func(opts *Options) {
		opts.debug = len(b) == 0 || b[0]
	}
}

func Func(name string, fn interface{}) Option {
	return func(opts *Options) {
		opts.fm[name] = fn
	}
}

func FuncMap(m map[string]interface{}) Option {
	return func(opts *Options) {
		for k, v := range m {
			opts.fm[k] = v
		}
	}
}

type Renderer struct {
	opts *Options
	t    *template.Template
}

func New(opts ...Option) (r *Renderer, err error) {
	options := &Options{
		fm: make(template.FuncMap),
	}
	for _, opt := range opts {
		opt(options)
	}
	err = options.ensure()
	if err != nil {
		return
	}

	r = &Renderer{
		opts: options,
	}
	err = r.compile()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func Must(opts ...Option) *Renderer {
	r, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, ctx web.Context) (err error) {
	return r.t.ExecuteTemplate(w, name, data)
}

func (r *Renderer) compile() (err error) {
	if r.opts.fs == nil {
		err = filepath.Walk(r.opts.dir, func(p string, d os.FileInfo, e error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			ext := filepath.Ext(p)
			if !r.match(ext) {
				return nil
			}

			rel, err := filepath.Rel(r.opts.dir, p)
			if err != nil {
				return err
			}

			data, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}

			name := filepath.ToSlash(rel[0 : len(rel)-len(ext)])
			r.parse(name, data)
			return nil
		})
	} else {
		err = fs.WalkDir(r.opts.fs, r.opts.dir, func(p string, d fs.DirEntry, e error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			ext := filepath.Ext(p)
			if !r.match(ext) {
				return nil
			}

			rel, err := filepath.Rel(r.opts.dir, p)
			if err != nil {
				return err
			}

			data, err := fs.ReadFile(r.opts.fs, p)
			if err != nil {
				return err
			}

			name := filepath.ToSlash(rel[0 : len(rel)-len(ext)])
			r.parse(name, data)
			return nil
		})
	}
	return
}

func (r *Renderer) match(ext string) bool {
	for _, e := range r.opts.exts {
		if e == ext {
			return true
		}
	}
	return false
}

func (r *Renderer) parse(name string, data []byte) {
	var tpl *template.Template
	if r.t == nil {
		r.t = template.New(name)
		tpl = r.t
	} else {
		tpl = r.t.New(name)
	}
	tpl.Funcs(r.opts.fm)
	template.Must(tpl.Parse(string(data)))
}

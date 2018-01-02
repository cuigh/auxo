package std

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/net/web"
)

type Options struct {
	debug bool
	dir   string
	exts  []string
	fm    template.FuncMap
}

func (opts *Options) ensure() error {
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
	if len(opts.exts) == 0 {
		opts.exts = []string{".html", ".gohtml"}
	}
	return nil
}

type Option func(opts *Options)

func Dir(dir string) Option {
	return func(opts *Options) {
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
	if !r.opts.debug {
		r.t, err = r.compile()
		if err != nil {
			return nil, err
		}
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
	var t *template.Template
	if r.opts.debug {
		// on debug mode, always recompile templates
		t, err = r.compile()
		if err != nil {
			return
		}
	} else {
		t = r.t
	}
	return t.ExecuteTemplate(w, name, data)
}

func (r *Renderer) compile() (*template.Template, error) {
	t := template.New(r.opts.dir)
	// Walk the supplied directory and compile any files that match our extension list.
	if err := filepath.Walk(r.opts.dir, func(path string, info os.FileInfo, err error) error {
		// fmt.Println("path: ", path)
		// if is a dir, return immediately.(dir is not a valid golang template)
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(r.opts.dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}

		for _, e := range r.opts.exts {
			if ext == e {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := rel[0 : len(rel)-len(ext)]
				// fmt.Println("name: ", filepath.ToSlash(name))
				tpl := t.New(filepath.ToSlash(name))
				tpl.Funcs(r.opts.fm)

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tpl.Parse(string(buf)))
				break
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return t, nil
}

package std

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cuigh/auxo/net/web"
)

type Renderer struct {
	dir        string
	extensions []string
	debug      bool
	tree       *template.Template
	funcs      template.FuncMap
	locker     sync.Locker
}

func New(dir string) (r *Renderer, err error) {
	r = &Renderer{
		dir:        dir,
		extensions: []string{".html", ".gohtml"},
	}
	r.tree, err = r.compile()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func Must(dir string) *Renderer {
	r, err := New(dir)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Renderer) SetDebug(b bool) *Renderer {
	r.debug = b
	return r
}

func (r *Renderer) SetExtensions(ext ...string) *Renderer {
	if len(ext) == 0 {
		ext = []string{".html", ".gohtml"}
	}
	r.extensions = ext
	return r
}

func (r *Renderer) AddFunc(name string, fn interface{}) *Renderer {
	r.funcs[name] = fn
	return r
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, ctx web.Context) error {
	// on debug mode, always recompile templates
	if r.debug {
		tree, err := r.compile()
		if err != nil {
			return err
		}
		return tree.ExecuteTemplate(w, name, data)
	}

	if r.tree != nil {
		return r.tree.ExecuteTemplate(w, name, data)
	}

	r.locker.Lock()
	defer r.locker.Unlock()

	if r.tree == nil {
		tree, err := r.compile()
		if err != nil {
			return err
		}
		r.tree = tree
	}
	return r.tree.ExecuteTemplate(w, name, data)
}

func (r *Renderer) compile() (*template.Template, error) {
	tree := template.New(r.dir)
	// Walk the supplied directory and compile any files that match our extension list.
	if err := filepath.Walk(r.dir, func(path string, info os.FileInfo, err error) error {
		// fmt.Println("path: ", path)
		// if is a dir, return immediately.(dir is not a valid golang template)
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(r.dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}

		for _, extension := range r.extensions {
			if ext == extension {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := rel[0 : len(rel)-len(ext)]
				// fmt.Println("name: ", filepath.ToSlash(name))
				tpl := tree.New(filepath.ToSlash(name))
				tpl.Funcs(r.funcs)

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tpl.Parse(string(buf)))
				break
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return tree, nil
}

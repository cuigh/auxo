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

//var VariableKey = "$jet$"

type Renderer struct {
	set   *jet.Set
	trans jet.Translator
}

type Config struct {
	Debug bool
}

func New(dir ...string) *Renderer {
	if len(dir) == 0 {
		d := filepath.Dir(app.GetPath())
		p := filepath.Join(d, "views")
		if files.Exist(p) {
			dir = append(dir, p)
		} else {
			p = filepath.Join(d, "resources/views")
			if files.Exist(p) {
				dir = append(dir, p)
			}
		}
	}
	if len(dir) == 0 {
		panic(errors.New("jet: templates directory is missing."))
	}

	r := &Renderer{
		set: jet.NewHTMLSet(dir...),
	}
	// Add common functions
	r.AddFunc("printf", fmt.Sprintf)
	r.AddFunc("limit", renderer.Limit)
	r.AddFunc("slice", renderer.Slice)
	return r
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, ctx web.Context) error {
	tpl, err := r.set.GetTemplate(name)
	if err == nil {
		//var variables = ctx.Get(VariableKey)
		//trans := ctx.Get("$translator")
		//err = tpl.ExecuteI18N(r.trans, w, variables, data)
		err = tpl.Execute(w, nil, data)
	}
	return err
}

func (r *Renderer) SetDebug(b bool) *Renderer {
	r.set.SetDevelopmentMode(b)
	return r
}

//func (r *Renderer) SetTranslator(translator jet.Translator) *Renderer {
//	r.trans = translator
//	return r
//}

func (r *Renderer) AddFuncs(fns map[string]interface{}) *Renderer {
	for name, fn := range fns {
		r.set.AddGlobal(name, fn)
	}
	return r
}

func (r *Renderer) AddFunc(name string, fn interface{}) *Renderer {
	r.set.AddGlobal(name, fn)
	return r
}

// add fast func
//func (r *Renderer) AddFunc(name string, fn jet.Func) *Renderer {
//	r.set.AddGlobalFunc(name, fn)
//	return r
//}

func (r *Renderer) AddVariable(name string, value interface{}) *Renderer {
	r.set.AddGlobal(name, value)
	return r
}

func (r *Renderer) AddVariables(vars map[string]interface{}) *Renderer {
	for name, value := range vars {
		r.set.AddGlobal(name, value)
	}
	return r
}

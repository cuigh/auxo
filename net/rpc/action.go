package rpc

import (
	"reflect"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/util/debug"
)

var newers = map[reflect.Type]Newer{}

func RegisterNewer(t reflect.Type, n Newer) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	newers[t] = n
}

// Newer is a function to generate a value instead of reflect.New.
type Newer func(t reflect.Type) interface{}

func defaultNewer(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

type SHandler func(c Context) (interface{}, error)

type SFilter func(SHandler) SHandler

type ActionSet interface {
	Get(name string) Action
	Find(service, method string) Action
	Range(fn func(a Action) bool)
}

type actionSet struct {
	actions map[string]Action
}

func newActionSet() *actionSet {
	return &actionSet{
		actions: make(map[string]Action),
	}
}

func (s *actionSet) Get(name string) Action {
	return s.actions[name]
}

func (s *actionSet) Find(service, method string) Action {
	return s.Get(service + "." + method)
}

func (s *actionSet) Range(fn func(a Action) bool) {
	for _, action := range s.actions {
		if !fn(action) {
			return
		}
	}
}

func (s *actionSet) RegisterFunc(service, method string, fn interface{}, filter ...SFilter) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return errors.New("fn must be a function")
	}
	s.registerAction(service, method, v, filter...)
	return nil
}

func (s *actionSet) RegisterService(name string, svc interface{}, filter ...SFilter) error {
	if name == "" {
		name = reflect.TypeOf(svc).Name()
	}
	logger := log.Get(PkgName)

	v := reflect.ValueOf(svc)
	t := v.Type()
	for i, num := 0, t.NumMethod(); i < num; i++ {
		mi := t.Method(i)
		if mi.PkgPath != "" { // Method must be exported.
			continue
		}

		if out := mi.Type.NumOut(); out < 2 {
			s.registerAction(name, mi.Name, v.Method(i), filter...)
		} else if out == 2 {
			if mi.Type.Out(1) == reflects.TypeError {
				s.registerAction(name, mi.Name, v.Method(i), filter...)
			} else {
				logger.Debugf("skip method %s.%s: the second out arg must be error", name, mi.Name)
			}
		} else {
			logger.Debugf("skip method %s.%s: too much out args", name, mi.Name)
		}
	}

	sv := reflects.StructOf(reflect.ValueOf(svc))
	sv.VisitFields(func(fv reflect.Value, fi *reflect.StructField) error {
		if fv.Kind() == reflect.Func && fi.PkgPath == "" && !fv.IsNil() {
			s.registerAction(name, fi.Name, fv, filter...)
		}
		return nil
	})
	return nil
}

func (s *actionSet) registerAction(service, method string, fn reflect.Value, filter ...SFilter) {
	a := newAction(service, method, fn, filter...)
	s.actions[a.name] = a
	log.Get(PkgName).Debugf("register method: %s.%s", service, method)
}

// Action is the interface that wraps the methods of service executor.
type Action interface {
	// Name is the name of action, normally is '[Service].[Method]'
	Name() string
	// In returns input arguments of the action
	In() []reflect.Type
	// In returns output arguments of the action
	Out() []reflect.Type
	// Context returns true if the first in-arg is `context.Context`
	Context() bool
	// Error returns true if the last out-arg is `error`
	Error() bool
	// Handler returns real executor of the action
	Handler() SHandler
	fillArgs(c *context) (args []interface{})
	//do(c Context) (interface{}, error)
}

type action struct {
	name    string
	service string
	method  string
	handler SHandler
	v       reflect.Value
	in      []reflect.Type
	out     []reflect.Type
	args    []actionArg
	ctx     bool
	err     bool
}

func newAction(service, method string, v reflect.Value, filter ...SFilter) *action {
	a := &action{
		name:    service + "." + method,
		service: service,
		method:  method,
		v:       v,
	}

	a.handler = a.do
	for i := len(filter) - 1; i >= 0; i-- {
		a.handler = filter[i](a.handler)
	}

	t := v.Type()
	if n := t.NumIn(); n > 0 {
		a.in = make([]reflect.Type, n)
		a.args = make([]actionArg, n)
		for i := 0; i < n; i++ {
			a.in[i] = t.In(i)
			a.args[i] = newActionArg(t.In(i))
		}
		if a.in[0] == reflects.TypeContext {
			a.ctx = true
		}
	}
	if n := t.NumOut(); n > 0 {
		a.out = make([]reflect.Type, n)
		for i := 0; i < n; i++ {
			a.out[i] = t.Out(i)
		}
		if a.out[n-1] == reflects.TypeError {
			a.err = true
		}
	}
	return a
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Handler() SHandler {
	return a.handler
}

func (a *action) In() []reflect.Type {
	return a.in
}

func (a *action) Out() []reflect.Type {
	return a.out
}

func (a *action) Context() bool {
	return a.ctx
}

func (a *action) Error() bool {
	return a.err
}

func (a *action) fillArgs(c *context) (args []interface{}) {
	if l := len(a.in); l > 0 {
		c.req.Args = make([]interface{}, l)
		if a.ctx {
			c.req.Args[0] = c.ctx
			for i := 1; i < l; i++ {
				c.req.Args[i] = a.args[i].New()
			}
			return c.req.Args[1:]
		} else {
			for i := 0; i < l; i++ {
				c.req.Args[i] = reflect.New(a.in[i]).Interface()
			}
			return c.req.Args
		}
	}
	return
}

func (a *action) do(c Context) (r interface{}, err error) {
	// supported function types:
	//
	// type 1: func(...)
	// type 2: func(...) result
	// type 3: func(...) error
	// type 4: func(...) (result, error)
	// type 5: func(ctx, ...) (result, error)
	defer func() {
		if e := recover(); e != nil {
			log.Get(PkgName).Errorf("rpc > PANIC: %v, stack:\n%s", e, debug.Stack())
			err = errors.Convert(e)
		}
	}()

	var in []reflect.Value
	req := c.Request()
	if n := len(req.Args); n > 0 {
		in = make([]reflect.Value, n)
		for i, arg := range req.Args {
			in[i] = a.args[i].Value(arg)
		}
	}
	out := a.v.Call(in)
	switch len(out) {
	case 1:
		if !a.err {
			r = out[0].Interface()
		} else if !out[0].IsNil() {
			err = out[0].Interface().(error)
		}
	case 2:
		r = out[0].Interface()
		if !out[1].IsNil() {
			err = out[1].Interface().(error)
		}
	}
	return
}

type actionArg interface {
	New() interface{}
	Value(i interface{}) reflect.Value
}

func newActionArg(t reflect.Type) actionArg {
	switch t.Kind() {
	case reflect.Interface:
		return interfaceArg{}
	case reflect.Ptr:
		return ptrArg{t: t.Elem()}
	default:
		return valueArg{t: t}
	}
}

type interfaceArg struct {
}

func (interfaceArg) New() interface{} {
	panic(errors.NotSupported)
}

func (interfaceArg) Value(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}

type ptrArg struct {
	t reflect.Type
	//n Newer
}

func (arg ptrArg) New() interface{} {
	//return arg.n(arg.t)
	return reflect.New(arg.t).Interface()
}

func (arg ptrArg) Value(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}

type valueArg struct {
	t reflect.Type
	//n Newer
}

func (arg valueArg) New() interface{} {
	//return arg.n(arg.t)
	return reflect.New(arg.t).Interface()
}

func (arg valueArg) Value(i interface{}) reflect.Value {
	return reflect.ValueOf(i).Elem()
}

package ioc

import (
	"reflect"
	"unsafe"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/util/lazy"
)

var global = New()

// TODO: generic method for golang v1.18
//func Get[T any]() T {
//	return global.Get[T]()
//}

func Find[T any](name string) T {
	return global.Find(name).(T)
}

func TryFind[T any](name string) (t T, err error) {
	var s any
	s, err = global.TryFind(name)
	if err == nil {
		t = s.(T)
	}
	return
}

func Range(fn func(name string, service any) bool) {
	global.Range(fn)
}

func Put(builder any, opts ...Option) {
	global.Put(builder, opts...)
}

func Call(fn any) (err error) {
	return global.Call(fn)
}

func Bind(target any) (err error) {
	return global.Bind(target)
}

type Container struct {
	names map[string]*service
	types map[reflect.Type]*service
}

func New() *Container {
	return &Container{
		names: make(map[string]*service),
		types: make(map[reflect.Type]*service),
	}
}

// TryFind returns the service instance with the specified name.
func (c *Container) TryFind(name string) (any, error) {
	s, ok := c.names[name]
	if !ok {
		return nil, errors.Format("container: service '%s' not registered", name)
	}
	return s.instance(c)
}

// Find returns the service instance with the specified name.
func (c *Container) Find(name string) any {
	s, err := c.TryFind(name)
	if err != nil {
		panic(err)
	}
	return s
}

// Range calls fn sequentially for each service present in the map. If fn returns false, range stops the iteration.
func (c *Container) Range(fn func(name string, service any) bool) {
	for n, s := range c.names {
		i, err := s.instance(c)
		if err != nil {
			panic(err)
		}
		if !fn(n, i) {
			return
		}
	}
}

// Put registers the service to container.
func (c *Container) Put(builder any, opts ...Option) {
	t := reflect.TypeOf(builder)
	v := reflect.ValueOf(builder)
	if t.Kind() != reflect.Func || t.NumOut() != 1 || v.IsNil() {
		panic("container: builder must be a function with one return value")
	}

	s := &service{t: t, v: v, singleton: true}
	for _, opt := range opts {
		opt(s)
	}
	if s.singleton {
		s.value = &lazy.Value[any]{
			New: func() (any, error) {
				return s.build(c)
			},
		}
	}

	if s.name != "" {
		if _, ok := c.names[s.name]; ok {
			panic("duplicate service name: " + s.name)
		} else {
			c.names[s.name] = s
		}
	}
	if _, ok := c.types[t.Out(0)]; !ok {
		c.types[t.Out(0)] = s
	}
}

// Call invoke fn with specified services as it's params.
func (c *Container) Call(fn any) error {
	v := reflect.ValueOf(fn)
	t := reflect.TypeOf(fn)
	if v.IsNil() || t.Kind() != reflect.Func {
		return errors.New("container: argument 'fn' is not a valid function")
	}

	results, err := c.call(t, v)
	if err != nil {
		return err
	} else if len(results) > 0 && results[0].Type() == reflects.TypeError && !results[0].IsNil() {
		return results[0].Interface().(error)
	}
	return nil
}

func (c *Container) call(t reflect.Type, v reflect.Value) ([]reflect.Value, error) {
	args := make([]reflect.Value, t.NumIn())
	for i := range args {
		argType := t.In(i)
		svc, err := c.get(argType)
		if err != nil {
			return nil, err
		}
		args[i] = reflect.ValueOf(svc)
	}
	return v.Call(args), nil
}

// Bind fills struct fields with specific services.
func (c *Container) Bind(target any) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		elem := targetType.Elem()
		if elem.Kind() == reflect.Struct {
			targetValue := reflect.ValueOf(target).Elem()
			targetType := targetValue.Type()
			for i := 0; i < targetValue.NumField(); i++ {
				tag, exist := targetType.Field(i).Tag.Lookup("container")
				if !exist {
					continue
				}

				var (
					f   = targetValue.Field(i)
					svc any
					err error
				)
				switch tag {
				case "type":
					svc, err = c.get(f.Type())
				case "name":
					svc, err = c.TryFind(targetType.Field(i).Name)
				default:
					err = errors.Format("container: tag of '%v' field is invalid", targetType.Field(i).Name)
				}

				if err != nil {
					return err
				}

				ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
				ptr.Set(reflect.ValueOf(svc))
			}
			return nil
		}
	}

	return errors.New("container: target must be a pointer of structure")
}

func (c *Container) get(t reflect.Type) (any, error) {
	if s := c.types[t]; s != nil {
		return s.instance(c)
	}
	return nil, errors.Format("container: cannot resolve service '%s'", t)
}

type service struct {
	t     reflect.Type
	v     reflect.Value
	value *lazy.Value[any]

	// options
	name      string
	group     string
	singleton bool
}

func (s *service) instance(c *Container) (any, error) {
	if s.singleton {
		return s.value.Get()
	}
	return s.build(c)
}

func (s *service) build(c *Container) (any, error) {
	results, err := c.call(s.t, s.v)
	if err != nil {
		return nil, err
	}
	return results[0].Interface(), nil
}

type Option func(opts *service)

// Name set service's name.
func Name(name string) Option {
	return func(opts *service) {
		opts.name = name
	}
}

// Group set service's group.
func Group(group string) Option {
	return func(opts *service) {
		opts.group = group
	}
}

// Scope set service's scope.
func Scope(singleton bool) Option {
	return func(opts *service) {
		opts.singleton = singleton
	}
}

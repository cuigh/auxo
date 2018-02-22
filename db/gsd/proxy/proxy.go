package proxy

import (
	"reflect"
	"strings"

	"github.com/cuigh/auxo/cache"
	_ "github.com/cuigh/auxo/cache/memory"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/log"
)

const PkgName = "auxo.db.gsd.proxy"

var (
	Default = New(nil)
)

type Builder func(db *gsd.LazyDB, ft reflect.Type, options data.Options) reflect.Value

func Register(action string, builder Builder) {
	Default.Register(action, builder)
}

func ApplyLazy(db *gsd.LazyDB, i interface{}) {
	Default.ApplyLazy(db, i)
}

// Apply generate data access methods according to fields of i.
func Apply(dbName string, i interface{}) {
	Default.Apply(dbName, i)
}

type Options struct {
	//Before func(ins []reflect.Value)
	//After  func(ins []reflect.Value)
}

type Proxy struct {
	options  *Options
	builders map[string]Builder
}

func New(options *Options) *Proxy {
	if options == nil {
		options = &Options{}
	}
	p := &Proxy{
		options: options,
	}
	// default builders
	p.builders = map[string]Builder{
		"load":   buildLoadProxy,
		"find":   buildFindProxy,
		"remove": buildRemoveProxy,
		"modify": buildModifyProxy,
		"create": buildCreateProxy,
		//"delete": buildDeleteProxy,
	}
	return p
}

func (p *Proxy) Register(action string, builder Builder) {
	p.builders[action] = builder
}

// Apply generate data access methods according to fields of i.
func (p *Proxy) Apply(dbName string, i interface{}) {
	p.ApplyLazy(gsd.Lazy(dbName), i)
}

func (p *Proxy) ApplyLazy(db *gsd.LazyDB, i interface{}) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		panic("i must be struct pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic("i must be func pointer")
	}

	t := v.Type()
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := v.Field(i)
		if f.Kind() != reflect.Func {
			continue
		}

		sf := t.Field(i)
		tag := sf.Tag.Get("gsd") // gsd:"select,table:user"
		if tag == "" {
			continue
		}
		action, options := p.parseTag(tag)
		fn := p.buildProxy(db, action, options, sf.Type)
		f.Set(fn)
	}
}

func (p *Proxy) parseTag(tag string) (action string, opts data.Options) {
	// gsd:"select,table:user"
	items := strings.SplitN(tag, ",", 2)
	action = strings.ToLower(items[0])
	if len(items) > 1 {
		opts = data.Options{}
		for i := 1; i < len(items); i++ {
			kv := strings.SplitN(items[i], ":", 2)
			if len(kv) == 2 {
				opts = append(opts, data.Option{Name: kv[0], Value: kv[1]})
			} else {
				log.Get(PkgName).Warn("invalid option: ", items[i])
			}
		}
	}
	return
}

func (p *Proxy) buildProxy(db *gsd.LazyDB, action string, options data.Options, ft reflect.Type) (v reflect.Value) {
	if b := p.builders[action]; b != nil {
		return b(db, ft, options)
	}
	panic("not supported action: " + action)
}

func buildFindProxy(lazy *gsd.LazyDB, ft reflect.Type, options data.Options) (v reflect.Value) {
	// todo: validate func
	rt, ret := ft.Out(0), ft.Out(0) // result type / result elem type
	if ret.Kind() == reflect.Ptr {
		ret = ret.Elem()
	}
	zr := reflect.Zero(rt) // zero result
	key := options.Get("cache")
	if key == "" {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 2)

			db, err := lazy.Try()
			if err != nil {
				outs[0] = zr
				outs[1] = reflects.Error(err)
				return
			}

			m := gsd.GetMeta(rt)
			w := &gsd.SimpleCriteriaSet{}
			for i, key := range m.PrimaryKeys {
				w.Equal(key, ins[i].Interface())
			}
			r := reflect.New(ret)
			err = db.Select(m.Selects...).From(m.Table).Where(w).Fill(r.Interface())
			if err != nil {
				outs[0] = zr
				outs[1] = reflects.Error(err)
				return
			}

			if rt == ret {
				outs[0] = r.Elem()
			} else {
				outs[0] = r
			}
			outs[1] = reflects.ZeroError
			return
		})
	} else {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 2)
			var (
				r    = reflect.New(ret)
				i    = r.Interface()
				m    = gsd.GetMeta(rt)
				args = make([]interface{}, len(ins))
			)
			for i, in := range ins {
				args[i] = in.Interface()
			}

			// check cache
			if value := cache.Get(key, args...); value != nil {
				if err := value.Scan(i); err == nil {
					if rt == ret {
						outs[0] = r.Elem()
					} else {
						outs[0] = r
					}
					outs[1] = reflects.ZeroError
					return
				} else {
					log.Get(PkgName).Debug("load from cache failed: ", err)
				}
			}

			db, err := lazy.Try()
			if err == nil {
				w := &gsd.SimpleCriteriaSet{}
				for i, key := range m.PrimaryKeys {
					w.Equal(key, args[i])
				}
				err = db.Select(m.Selects...).From(m.Table).Where(w).Fill(i)
			}
			if err != nil {
				outs[0], outs[1] = zr, reflects.Error(err)
				return
			}

			if rt == ret {
				outs[0] = r.Elem()
			} else {
				outs[0] = r
			}
			outs[1] = reflects.ZeroError
			cache.Set(i, key, args...)
			return
		})
	}
}

func buildLoadProxy(lazy *gsd.LazyDB, ft reflect.Type, options data.Options) (v reflect.Value) {
	// todo: validate func
	key := options.Get("cache")
	if key == "" {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 1)
			db, err := lazy.Try()
			if err == nil {
				err = db.Load(ins[0].Interface())
			}
			outs[0] = reflects.Error(err)
			return
		})
	} else {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 1)
			var (
				i    = ins[0].Interface()
				m    = gsd.GetMeta(ins[0].Type())
				args = m.PrimaryKeyValues(i)
			)

			// check cache
			if value := cache.Get(key, args...); value != nil {
				if err := value.Scan(i); err == nil {
					outs[0] = reflects.ZeroError
					return
				} else {
					log.Get(PkgName).Debug("load from cache failed: ", err)
				}
			}

			db, err := lazy.Try()
			if err == nil {
				err = db.Load(i)
				if err == nil {
					cache.Set(i, key, args...)
				}
			}
			outs[0] = reflects.Error(err)
			return
		})
	}
}

func buildRemoveProxy(lazy *gsd.LazyDB, ft reflect.Type, _ data.Options) (v reflect.Value) {
	// todo: validate func
	return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
		outs = make([]reflect.Value, 1)

		db, err := lazy.Try()
		if err == nil {
			_, err = db.Remove(ins[0].Interface())
		}
		outs[0] = reflects.Error(err)
		return
	})
}

func buildModifyProxy(lazy *gsd.LazyDB, ft reflect.Type, _ data.Options) (v reflect.Value) {
	// todo: validate func
	return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
		outs = make([]reflect.Value, 1)

		db, err := lazy.Try()
		if err == nil {
			_, err = db.Modify(ins[0].Interface())
		}
		outs[0] = reflects.Error(err)
		return
	})
}

func buildCreateProxy(lazy *gsd.LazyDB, ft reflect.Type, _ data.Options) (v reflect.Value) {
	// todo: validate func
	return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
		outs = make([]reflect.Value, 1)

		db, err := lazy.Try()
		if err == nil {
			err = db.Create(ins[0].Interface())
		}
		outs[0] = reflects.Error(err)
		return
	})
}

func buildSearchProxy(lazy *gsd.LazyDB, ft reflect.Type, options data.Options) (v reflect.Value) {
	//func(f *UserFilter) ([]*User, int, error) `gsd:"search,cache:user.search"`
	// todo: validate func
	key := options.Get("cache")
	if key == "" {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 1)
			db, err := lazy.Try()
			if err == nil {
				var (
					i     = ins[0].Interface()
					count int
				)
				fm := gsd.GetMeta(ins[0].Type())
				rm := gsd.GetMeta(outs[0].Elem().Type())
				w := fm.WhereValues(i)
				err = db.Select(rm.Selects...).From(rm.Table).Where(w).Page(1, 1).List(nil, &count)
			}
			// todo:
			outs[2] = reflects.Error(err)
			return
		})
	} else {
		return reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
			outs = make([]reflect.Value, 1)
			var (
				i    = ins[0].Interface()
				m    = gsd.GetMeta(ins[0].Type())
				args = m.PrimaryKeyValues(i)
			)

			// check cache
			if value := cache.Get(key, args...); value != nil {
				if err := value.Scan(i); err == nil {
					outs[0] = reflects.ZeroError
					return
				} else {
					log.Get(PkgName).Debug("load from cache failed: ", err)
				}
			}

			db, err := lazy.Try()
			if err == nil {
				err = db.Load(i)
				if err == nil {
					cache.Set(i, key, args...)
				}
			}
			outs[0] = reflects.Error(err)
			return
		})
	}
}

package cache

import (
	"time"

	"sync"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/util/lazy"
)

const PkgName = "auxo.cache"

var (
	providers = make(map[string]func(opts data.Map) (Provider, error))
	f         = new(factory)
	def       = lazy.Value{
		New: func() (interface{}, error) {
			c, err := GetCacher("")
			if err != nil {
				log.Get(PkgName).Warn("cache > initialize default Cacher failed: ", err)
				c = &cacher{enabled: false}
			}
			return c, nil
		},
	}
)

// Provider is cache provider interface.
type Provider interface {
	// Get returns cached value, provider should return data.Nil instead of nil when cache is invalid.
	Get(key string) (data.Value, error)
	Set(key string, value interface{}, expiry time.Duration) error
	Remove(key string) error
	Exist(key string) (bool, error)
}

func Register(name string, f func(opts data.Map) (Provider, error)) {
	providers[name] = f
}

func GetCacher(name string) (Cacher, error) {
	return f.GetCacher(name)
}

func defaultCacher() *cacher {
	c, _ := def.Get()
	return c.(*cacher)
}

// Get returns cached value, the result is assured of not nil.
func Get(key string, args ...interface{}) data.Value {
	return defaultCacher().Get(key, args...)
}

func Set(value interface{}, key string, args ...interface{}) {
	defaultCacher().Set(value, key, args...)
}

func Exist(key string, args ...interface{}) bool {
	return defaultCacher().Exist(key, args...)
}

func Remove(key string, args ...interface{}) {
	defaultCacher().Remove(key, args...)
}

func RemoveGroup(key string) {
	defaultCacher().RemoveGroup(key)
}

type factory struct {
	locker  sync.Mutex
	cachers map[string]Cacher
}

func (f *factory) GetCacher(name string) (Cacher, error) {
	if c, ok := f.cachers[name]; ok {
		return c, nil
	}

	return f.initCacher(name)
}

func (f *factory) initCacher(name string) (Cacher, error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	if c, ok := f.cachers[name]; ok {
		return c, nil
	}

	if config.Exist("cache") {
		list := make([]*Options, 0)
		err := config.UnmarshalOption("cache", &list)
		if err != nil {
			return nil, err
		}

		for _, opts := range list {
			if opts.Name == name {
				pb := providers[opts.Provider]
				if pb == nil {
					return nil, errors.New("unknown cache provider: " + opts.Provider)
				}

				p, err := pb(opts.Options)
				if err != nil {
					return nil, err
				}

				c := newCacher(p, opts)
				// rebuild map to avoid locker
				cachers := make(map[string]Cacher)
				for k, v := range f.cachers {
					cachers[k] = v
				}
				cachers[name] = c
				f.cachers = cachers
				return c, nil
			}
		}
	}
	if name == "" {
		return nil, errors.New("can't find default cache config")
	} else {
		return nil, errors.New("can't find cache config for: " + name)
	}
}

func (f *factory) loadOptions(name string) (*Options, error) {
	key := "db.redis." + name
	if !config.Exist(key) {
		return nil, errors.Format("can't find redis config for [%s]", name)
	}

	opts := &Options{}
	err := config.UnmarshalOption(key, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

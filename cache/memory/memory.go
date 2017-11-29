package memory

import (
	"reflect"
	"sync"
	"time"

	"github.com/cuigh/auxo/cache"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
)

type item struct {
	value  interface{}
	expiry time.Time
}

func (i *item) IsNil() bool {
	return i.value == nil
}

func (i *item) Scan(value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Convert(e)
		}
	}()

	rv := reflect.ValueOf(i.value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	reflect.ValueOf(value).Elem().Set(rv)
	return
}

func (i *item) Bytes() ([]byte, error) {
	if v, ok := i.value.([]byte); ok {
		return v, nil
	}
	return nil, i.typeError("[]byte")
}

func (i *item) Bool() (bool, error) {
	if v, ok := i.value.(bool); ok {
		return v, nil
	}
	return false, i.typeError("bool")
}

func (i *item) Int() (int, error) {
	if v, ok := i.value.(int); ok {
		return v, nil
	}
	return 0, i.typeError("int")
}

func (i *item) Int8() (int8, error) {
	if v, ok := i.value.(int8); ok {
		return v, nil
	}
	return 0, i.typeError("int8")
}

func (i *item) Int16() (int16, error) {
	if v, ok := i.value.(int16); ok {
		return v, nil
	}
	return 0, i.typeError("int16")
}

func (i *item) Int32() (int32, error) {
	if v, ok := i.value.(int32); ok {
		return v, nil
	}
	return 0, i.typeError("int32")
}

func (i *item) Int64() (int64, error) {
	if v, ok := i.value.(int64); ok {
		return v, nil
	}
	return 0, i.typeError("int64")
}

func (i *item) Uint() (uint, error) {
	if v, ok := i.value.(uint); ok {
		return v, nil
	}
	return 0, i.typeError("uint")
}

func (i *item) Uint8() (uint8, error) {
	if v, ok := i.value.(uint8); ok {
		return v, nil
	}
	return 0, i.typeError("uint8")
}

func (i *item) Uint16() (uint16, error) {
	if v, ok := i.value.(uint16); ok {
		return v, nil
	}
	return 0, i.typeError("uint16")
}

func (i *item) Uint32() (uint32, error) {
	if v, ok := i.value.(uint32); ok {
		return v, nil
	}
	return 0, i.typeError("uint32")
}

func (i *item) Uint64() (uint64, error) {
	if v, ok := i.value.(uint64); ok {
		return v, nil
	}
	return 0, i.typeError("uint64")
}

func (i *item) Float32() (float32, error) {
	if v, ok := i.value.(float32); ok {
		return v, nil
	}
	return 0, i.typeError("float32")
}

func (i *item) Float64() (float64, error) {
	if v, ok := i.value.(float64); ok {
		return v, nil
	}
	return 0, i.typeError("float64")
}

func (i *item) String() (string, error) {
	if v, ok := i.value.(string); ok {
		return v, nil
	}
	return "", i.typeError("string")
}

func (i *item) typeError(t string) error {
	return errors.Format("type is %T, not %s", i.value, t)
}

func (i *item) Valid() bool {
	return i.expiry.After(time.Now())
}

// Provider is memory provider implementation.
type Provider struct {
	locker sync.RWMutex
	items  map[string]*item
}

func NewProvider() *Provider {
	p := &Provider{
		items: make(map[string]*item),
	}
	go p.removeExpired()
	return p
}

func (p *Provider) Get(key string) (value data.Value, err error) {
	p.locker.RLock()
	item, ok := p.items[key]
	p.locker.RUnlock()

	if ok && item.Valid() {
		return item, nil
	}
	return data.Nil, nil
}

func (p *Provider) Set(key string, value interface{}, expiry time.Duration) error {
	p.locker.Lock()
	p.items[key] = &item{
		value:  value,
		expiry: time.Now().Add(expiry),
	}
	p.locker.Unlock()
	return nil
}

func (p *Provider) Remove(key string) error {
	p.locker.Lock()
	delete(p.items, key)
	p.locker.Unlock()
	return nil
}

func (p *Provider) Exist(key string) (bool, error) {
	p.locker.RLock()
	item, ok := p.items[key]
	p.locker.RUnlock()
	return ok && item.expiry.After(time.Now()), nil
}

func (p *Provider) removeExpired() {
	for {
		time.Sleep(time.Minute * 10)

		p.locker.Lock()
		var keys []string
		for key, item := range p.items {
			if !item.Valid() {
				keys = append(keys, key)
			}
		}
		for _, key := range keys {
			delete(p.items, key)
		}
		p.locker.Unlock()
	}
}

func init() {
	cache.Register("memory", func(opts data.Map) (cache.Provider, error) {
		return NewProvider(), nil
	})
}

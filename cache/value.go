package cache

import (
	"sync"
	"time"
)

// Value is a simple auto refresh cache hold.
type Value struct {
	locker sync.Mutex
	value  interface{}
	next   time.Time
	TTL    time.Duration
	Load   func() (interface{}, error)
}

// Get return cached value, it will return expired value if dirty is true and loading failed.
func (v *Value) Get(dirty ...bool) (value interface{}, err error) {
	if v.value != nil && time.Now().Before(v.next) {
		return v.value, nil
	}

	v.locker.Lock()
	defer v.locker.Unlock()

	if v.value != nil && time.Now().Before(v.next) {
		return v.value, nil
	}

	value, err = v.Load()
	if err == nil {
		ttl := v.TTL
		if ttl == 0 {
			ttl = time.Hour
		}
		v.value, v.next = value, time.Now().Add(ttl)
	} else {
		if v.value != nil && (len(dirty) > 0 && dirty[0]) {
			return v.value, nil
		}
	}
	return
}

// MustGet return cached value, it panics if error occurs.
func (v *Value) MustGet(dirty ...bool) (value interface{}) {
	value, err := v.Get(dirty...)
	if err != nil {
		panic(err)
	}
	return value
}

// Reset clears internal cache value.
func (v *Value) Reset() {
	v.value = nil
}

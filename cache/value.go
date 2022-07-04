package cache

import (
	"sync"
	"time"
)

// Value is a simple auto refresh cache hold.
type Value[T any] struct {
	locker sync.Mutex
	value  interface{}
	next   time.Time
	TTL    time.Duration
	Load   func() (T, error)
}

// Get return cached value, it will return expired value if dirty is true and loading failed.
func (v *Value[T]) Get(dirty ...bool) (value T, err error) {
	if v.value != nil && time.Now().Before(v.next) {
		return v.value.(T), nil
	}

	v.locker.Lock()
	defer v.locker.Unlock()

	if v.value != nil && time.Now().Before(v.next) {
		return v.value.(T), nil
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
			return v.value.(T), nil
		}
	}
	return
}

// MustGet return cached value, it panics if error occurs.
func (v *Value[T]) MustGet(dirty ...bool) (value T) {
	value, err := v.Get(dirty...)
	if err != nil {
		panic(err)
	}
	return value
}

// Reset clears internal cache value.
func (v *Value[T]) Reset() {
	v.value = nil
}

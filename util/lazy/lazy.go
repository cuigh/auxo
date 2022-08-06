// Package lazy providers a lazy singleton implementation.
package lazy

import (
	"sync"
	"sync/atomic"
)

type Value[T any] struct {
	locker sync.Mutex
	value  T
	ok     int32
	New    func() (T, error)
}

func (l *Value[T]) Get() (value T, err error) {
	if atomic.LoadInt32(&l.ok) == 1 {
		return l.value, nil
	}

	l.locker.Lock()
	defer l.locker.Unlock()

	if l.ok == 0 {
		l.value, err = l.New()
		if err == nil {
			atomic.StoreInt32(&l.ok, 1)
		}
	}
	return l.value, err
}

// Package lazy providers a lazy singleton implementation.
package lazy

import (
	"sync"
)

type Value struct {
	locker sync.Mutex
	value  interface{}
	New    func() (interface{}, error)
}

func (l *Value) Get() (value interface{}, err error) {
	if l.value != nil {
		return l.value, nil
	}

	l.locker.Lock()
	defer l.locker.Unlock()

	if l.value == nil {
		l.value, err = l.New()
	}
	return l.value, err
}

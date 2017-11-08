package lazy

import "sync"

type Newer func() (interface{}, error)

type KeyNewer func(key interface{}) (interface{}, error)

type One struct {
	locker sync.Mutex
	value  interface{}
	newer  func() (interface{}, error)
}

func NewOne(newer Newer) *One {
	return &One{
		newer: newer,
	}
}

func (l *One) Get() (value interface{}, err error) {
	if l.value != nil {
		return l.value, nil
	}

	l.locker.Lock()
	defer l.locker.Unlock()

	if l.value == nil {
		l.value, err = l.newer()
	}
	return l.value, err
}

type Map struct {
	locker sync.RWMutex
	values map[interface{}]interface{}
	newer  KeyNewer
}

func NewMap(newer KeyNewer) *Map {
	return &Map{
		values: make(map[interface{}]interface{}),
		newer:  newer,
	}
}

func (m *Map) Get(key interface{}) (value interface{}, err error) {
	var ok bool

	m.locker.RLock()
	defer m.locker.RUnlock()

	if value, ok = m.values[key]; ok {
		return
	}

	m.locker.Lock()
	defer m.locker.Unlock()

	if value, ok = m.values[key]; !ok {
		value, err = m.newer(key)
		if err == nil {
			m.values[key] = value
		}
	}
	return
}

package breaker

import (
	"sync"
)

type Group struct {
	locker   sync.RWMutex
	breakers map[string]*Breaker
}

func NewGroup() *Group {
	return &Group{
		breakers: make(map[string]*Breaker),
	}
}

func (g *Group) Get(name string, builder func() *Breaker) *Breaker {
	g.locker.RLock()
	b := g.breakers[name]
	g.locker.RUnlock()

	if b != nil {
		return b
	}

	g.locker.Lock()
	defer g.locker.Unlock()

	b = g.breakers[name]
	if b == nil {
		b = builder()
		g.breakers[name] = b
	}
	return b
}

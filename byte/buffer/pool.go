package buffer

import (
	"sync"
)

type Pool struct {
	pool sync.Pool
}

func NewPool(size int) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, size)
			},
		},
	}
}

func (p *Pool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *Pool) Put(buf []byte) {
	p.pool.Put(buf)
}

type GroupPool struct {
	sizes []int
	pools []sync.Pool
}

func NewGroupPool(minSize, maxSize int, growFactor float32) *GroupPool {
	p := &GroupPool{}

	size := minSize
	for {
		p.sizes = append(p.sizes, size)
		p.pools = append(p.pools, p.createSyncPool(size))

		if size >= maxSize {
			break
		}
		size = int(float32(size) * growFactor)
	}

	return p
}

func (p *GroupPool) Get(size int) []byte {
	i := p.find(size)
	return p.pools[i].Get().([]byte)
}

func (p *GroupPool) Put(buf []byte) {
	i := p.find(cap(buf))
	p.pools[i].Put(buf)
}

func (p *GroupPool) createSyncPool(size int) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, size)
		},
	}
}

func (p *GroupPool) find(size int) int {
	// binary search, see: sort.Search(len(p.pools), func(i int) bool { return p.pools[i].size >= size })
	i, j := 0, len(p.sizes)
	for i < j {
		h := i + (j-i)/2 // avoid overflow when computing h
		// i ≤ h < j
		if p.sizes[h] < size {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	return i
}

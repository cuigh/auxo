package buffer

import (
	"fmt"
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestPool(t *testing.T) {
	p := NewPool(128)

	buf1 := p.Get()
	p.Put(buf1)

	buf2 := p.Get()
	p.Put(buf2)

	assert.Equal(t, fmt.Sprintf("%p", buf1), fmt.Sprintf("%p", buf2))
}

func TestGroupPool(t *testing.T) {
	p := NewGroupPool(128, 1024, 2)

	cases := []int{127, 128, 129, 130, 1024}
	for _, size := range cases {
		buf := p.Get(size)
		p.Put(buf)
		t.Logf("%p > size: %d, cap: %d", buf, size, cap(buf))
	}
}

func BenchmarkGroupPool(b *testing.B) {
	p := NewGroupPool(1<<10, 1<<20, 2)
	for i := 0; i < b.N; i++ {
		buf := p.Get(8192)
		p.Put(buf)
		//if len(buf) > 0 {
		//}
	}
}

//func BenchmarkNewBytes(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		buf := make([]byte, 0, 8192)
//		if len(buf) > 0 {
//		}
//	}
//}

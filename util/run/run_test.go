package run_test

import (
	"sync"
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/run"
)

func TestSafe(t *testing.T) {
	const s = "test"

	run.Safe(func() {
		panic(s)
	}, func(e interface{}) {
		assert.Equal(t, s, e)
	})
}

func TestCount(t *testing.T) {
	const s = "test"

	var g sync.WaitGroup
	run.Count(&g, func() {
		panic(s)
	}, func(e interface{}) {
		assert.Equal(t, s, e)
	})
}

func TestPool(t *testing.T) {
	p := run.Pool{
		Min: 1,
		Max: 5,
	}
	p.Start()
	for i := 0; i < 100; i++ {
		err := p.Put(func() {
			time.Sleep(2 * time.Millisecond)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	err := p.Wait(time.Second)
	t.Log(err)
	p.Stop()
}

package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/cuigh/auxo/apm/limiter"
	"github.com/cuigh/auxo/apm/limiter/store/memory"
	"github.com/cuigh/auxo/test/assert"
)

func TestStore_Get(t *testing.T) {
	s := memory.New()
	rate := &limiter.Rate{
		Period: time.Second,
		Count:  10,
	}

	for i := 0; i < 10; i++ {
		res, err := s.Get(context.TODO(), rate, "test")
		assert.NoError(t, err)
		assert.True(t, res.OK())
	}

	res, err := s.Get(context.TODO(), rate, "test")
	assert.NoError(t, err)
	assert.False(t, res.OK())
}

func TestStore_Peek(t *testing.T) {
	s := memory.New()
	rate := &limiter.Rate{
		Period: time.Second,
		Count:  10,
	}

	for i := 0; i < 10; i++ {
		res, err := s.Peek(context.TODO(), rate, "test")
		assert.NoError(t, err)
		assert.True(t, res.OK())
		assert.Equal(t, rate.Count, res.Remaining)
	}
}

func BenchmarkStore_Get(b *testing.B) {
	s := memory.New()
	rate := &limiter.Rate{
		Period: time.Second,
		Count:  1000000000,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Get(context.TODO(), rate, "test")
	}
}

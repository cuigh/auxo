package breaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cuigh/auxo/apm/breaker"
	"github.com/cuigh/auxo/test/assert"
)

var mockError = errors.New("mock error")

func TestBreaker_Fallback(t *testing.T) {
	mockResult := "fallback"

	var s string
	b := breaker.NewBreaker("test", breaker.ErrorCount(10), breaker.Options{})
	err := b.Try(func() error {
		return mockError
	}, func(err error) error {
		assert.Same(t, err, mockError)
		s = mockResult
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, mockResult, s)
}

func TestBreaker_Open(t *testing.T) {
	b := breaker.NewBreaker("test", breaker.ErrorCount(10), breaker.Options{})
	for i := 0; i < 9; i++ {
		b.Try(func() error {
			return mockError
		})
	}
	assert.Equal(t, breaker.Closed, b.State())

	err := b.Try(func() error {
		return mockError
	})
	assert.Same(t, mockError, err)
	assert.Equal(t, breaker.Open, b.State())

	err = b.Try(func() error {
		return mockError
	})
	assert.Same(t, breaker.ErrCircuitOpen, err)
}

func TestBreaker_HalfOpen(t *testing.T) {
	opts := breaker.Options{Window: time.Millisecond * 500}
	b := breaker.NewBreaker("test", breaker.ErrorCount(10), opts)
	for i := 0; i < 10; i++ {
		b.Try(func() error {
			return mockError
		})
	}
	assert.Equal(t, breaker.Open, b.State())

	time.Sleep(opts.Window)
	b.Try(func() error {
		return nil
	})
	assert.Equal(t, breaker.Closed, b.State())
}

func TestBreaker_Timeout(t *testing.T) {
	opts := breaker.Options{Timeout: time.Millisecond * 100}
	b := breaker.NewBreaker("test", breaker.ErrorCount(10), opts)
	err := b.Try(func() error {
		time.Sleep(opts.Timeout)
		return nil
	})
	assert.Equal(t, breaker.ErrTimeout, err)
}

func TestCondition(t *testing.T) {
	cases := []struct {
		cond    breaker.Condition
		success int
		failure int
		state   int32
	}{
		{breaker.ErrorRate(0.5, 10), 5, 5, breaker.Open},
		{breaker.ErrorRate(0.5, 10), 6, 4, breaker.Closed},
		{breaker.ErrorCount(10), 0, 10, breaker.Open},
		{breaker.ErrorCount(10), 1, 9, breaker.Closed},
		{breaker.ConsecutiveErrorCount(5), 5, 5, breaker.Open},
		{breaker.ConsecutiveErrorCount(5), 6, 4, breaker.Closed},
	}

	for _, c := range cases {
		b := breaker.NewBreaker("test", c.cond, breaker.Options{})

		for i := 0; i < c.success; i++ {
			b.Try(func() error {
				return nil
			})
		}
		assert.Equal(t, breaker.Closed, b.State())

		for i := 0; i < c.failure; i++ {
			b.Try(func() error {
				return mockError
			})
		}
		assert.Equal(t, c.state, b.State())
	}
}

func BenchmarkBreaker_Try(b *testing.B) {
	opts := breaker.Options{Window: time.Millisecond * 500}
	br := breaker.NewBreaker("test", breaker.ErrorCount(10), opts)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		br.Try(func() error {
			return nil
		})
	}
}

func BenchmarkBreaker_TryWithTimeout(b *testing.B) {
	opts := breaker.Options{Window: time.Millisecond * 500, Timeout: time.Second}
	br := breaker.NewBreaker("test", breaker.ErrorCount(10), opts)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		br.Try(func() error {
			return nil
		})
	}
}

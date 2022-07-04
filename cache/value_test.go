package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cuigh/auxo/cache"
	"github.com/cuigh/auxo/test/assert"
)

func TestValue(t *testing.T) {
	i := 0
	v := cache.Value[int]{
		TTL: 100 * time.Millisecond,
		Load: func() (int, error) {
			if i == 0 {
				i++
				return i, nil
			}
			return 0, errors.New("mock error")
		},
	}

	value, err := v.Get()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	time.Sleep(101 * time.Millisecond)

	value, err = v.Get()
	assert.Error(t, err)
	assert.Equal(t, 0, value)

	value, err = v.Get(true)
	assert.NoError(t, err)
	assert.NotNil(t, value)

	v.Reset()
}

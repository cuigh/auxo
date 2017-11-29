package cache_test

import (
	"testing"
	"time"

	"errors"

	"github.com/cuigh/auxo/cache"
	"github.com/cuigh/auxo/test/assert"
)

func TestValue(t *testing.T) {
	i := 0
	v := cache.Value{
		TTL: 100 * time.Millisecond,
		Load: func() (interface{}, error) {
			if i == 0 {
				i++
				return &struct{}{}, nil
			}
			return nil, errors.New("mock error")
		},
	}

	value, err := v.Get()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	time.Sleep(101 * time.Millisecond)

	value, err = v.Get()
	assert.Error(t, err)
	assert.Nil(t, value)

	value, err = v.Get(true)
	assert.NoError(t, err)
	assert.NotNil(t, value)

	v.Reset()
}

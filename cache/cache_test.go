package cache_test

import (
	"testing"

	"github.com/cuigh/auxo/cache"
	_ "github.com/cuigh/auxo/cache/memory"
	_ "github.com/cuigh/auxo/cache/redis"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/test/assert"
)

const key = "key"

func init() {
	config.AddFolder(".")
}

func TestCache(t *testing.T) {
	var (
		value  = 10
		actual int
		err    error
	)

	// Remove
	cache.Remove(key)

	// Exist: false
	assert.False(t, cache.Exist(key))

	// Get: nil
	assert.True(t, cache.Get(key).IsNil())

	// Set
	cache.Set(value, key)

	// Exist: true
	assert.True(t, cache.Exist(key))

	// Get: not nil
	v := cache.Get(key)
	assert.NotNil(t, v)

	// Get > Scan
	actual, err = v.Int()
	assert.NoError(t, err)
	assert.Equal(t, value, actual)

	// Remove
	cache.Remove(key)

	// Exist: false
	assert.False(t, cache.Exist(key))

	// Get: nil
	assert.True(t, cache.Get(key).IsNil())
}

func TestCacher(t *testing.T) {
	var (
		value  = 10
		actual int
		err    error
	)

	c, err := cache.GetCacher("redis")
	assert.NoError(t, err)

	// Remove
	c.Remove(key)

	// Exist: false
	assert.False(t, c.Exist(key))

	// Get: nil
	assert.True(t, c.Get(key).IsNil())

	// Set
	c.Set(value, key)

	// Exist: true
	assert.True(t, c.Exist(key))

	// Get: not nil
	v := c.Get(key)
	assert.NotNil(t, v)

	// Get > Scan
	actual, err = v.Int()
	assert.NoError(t, err)
	assert.Equal(t, value, actual)

	// Remove
	c.Remove(key)

	// Exist: false
	assert.False(t, c.Exist(key))

	// Get: nil
	assert.True(t, c.Get(key).IsNil())
}

func TestRemoveVersion(t *testing.T) {
	cache.Set(1, "test1")
	cache.Set(1, "test2", 1)
	assert.True(t, cache.Exist("test1"))
	assert.True(t, cache.Exist("test2", 1))

	cache.RemoveGroup("test_version")
	assert.False(t, cache.Exist("test1"))
	assert.False(t, cache.Exist("test2", 1))
}

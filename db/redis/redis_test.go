package redis_test

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/db/redis"
	"github.com/cuigh/auxo/test/assert"
)

func init() {
	config.AddFolder(".")
}

func TestFactory_Open(t *testing.T) {
	cmd, err := redis.Open("cache")
	assert.NoError(t, err)

	err = cmd.Ping().Err()
	assert.NoError(t, err)
}

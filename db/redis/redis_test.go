package redis

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/test/assert"
)

func TestFactory_Open(t *testing.T) {
	config.AddFolder(".")

	cmd, err := Open("cache")
	assert.NoError(t, err)

	err = cmd.Ping().Err()
	assert.NoError(t, err)
}

package data

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestChannel(t *testing.T) {
	c := make(Chan, 1)

	c.Send()
	assert.True(t, c.TryReceive())
	assert.False(t, c.TryReceive())

	assert.True(t, c.TrySend())
	assert.False(t, c.TrySend())
	close(c)
}

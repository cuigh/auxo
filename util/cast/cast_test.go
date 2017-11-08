package cast_test

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/cast"
)

func TestToBool(t *testing.T) {
	assert.True(t, cast.ToBool(true))
	assert.False(t, cast.ToBool(false))
	assert.True(t, cast.ToBool("true"))
	assert.False(t, cast.ToBool("false"))
	assert.False(t, cast.ToBool(0))
	assert.True(t, cast.ToBool(1))
	assert.True(t, cast.ToBool(-1))
}

func TestToInt(t *testing.T) {
	var i int = 1
	assert.Equal(t, 0, cast.ToInt(nil))
	assert.Equal(t, 1, cast.ToInt(true))
	assert.Equal(t, 0, cast.ToInt(false))
	assert.Equal(t, 1, cast.ToInt(1))
	assert.Equal(t, -1, cast.ToInt(-1))
	assert.Equal(t, 1, cast.ToInt(1.0))
	assert.Equal(t, 1, cast.ToInt("1"))
	assert.Equal(t, i, cast.ToInt(&i))
}

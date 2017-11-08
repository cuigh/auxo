package cast_test

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/cast"
)

func TestTryToBool(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expected bool
		Error    bool
	}{
		{true, true, false},
		{false, false, false},
		{"true", true, false},
		{"false", false, false},
		{1, true, false},
		{0, false, false},
		{"x", false, true},
	}

	for _, c := range cases {
		r, err := cast.TryToBool(c.Input)
		if c.Error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, c.Expected, r)
		}
	}
}

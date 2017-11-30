package limiter_test

import (
	"testing"

	"github.com/cuigh/auxo/apm/limiter"
	"github.com/cuigh/auxo/test/assert"
)

func TestParseRate(t *testing.T) {
	cases := []struct {
		s  string
		ok bool
	}{
		{"", false},
		{"100", false},
		{"abc/5s", false},
		{"100/x", false},
		{"100/5s", true},
		{"1000/1m", true},
	}

	for _, c := range cases {
		_, err := limiter.ParseRate(c.s)
		if c.ok {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

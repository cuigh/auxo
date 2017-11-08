package base62_test

import (
	"math"
	"testing"

	"github.com/cuigh/auxo/encoding/base62"
	"github.com/cuigh/auxo/test/assert"
)

func TestEncode(t *testing.T) {
	cases := []struct {
		n uint64
		s string
	}{
		{0, "0"},
		{35, "Z"},
		{math.MaxUint64, "LygHa16AHYF"},
	}

	for _, c := range cases {
		s := base62.Encode(c.n)
		assert.Equal(t, c.s, s)

		n, err := base62.Decode(c.s)
		assert.NoError(t, err)
		assert.Equal(t, c.n, n)
	}
}

func TestFixed(t *testing.T) {
	cases := []struct {
		n uint64
		s string
	}{
		{0, "00000000000"},
		{35, "0000000000Z"},
		{math.MaxUint64, "LygHa16AHYF"},
	}

	for _, c := range cases {
		s := base62.EncodeFixed(c.n)
		assert.Equal(t, c.s, s)

		n, err := base62.Decode(c.s)
		assert.NoError(t, err)
		assert.Equal(t, c.n, n)
	}
}

package numbers

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestIntMaxMin(t *testing.T) {
	cases := []struct {
		nums []int
		min  int
		max  int
	}{
		{[]int{1}, 1, 1},
		{[]int{1, 2, 3, 4, 5}, 1, 5},
	}

	for _, c := range cases {
		assert.Equal(t, c.min, MinInt(c.nums...))
		assert.Equal(t, c.max, MaxInt(c.nums...))
	}
}

func TestIntLimit(t *testing.T) {
	const (
		min = 0
		max = 10
	)
	cases := []struct {
		n        int
		expected int
	}{
		{5, 5},
		{-1, 0},
		{11, 10},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, LimitInt(c.n, min, max))
	}
}

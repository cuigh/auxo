package set

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestSet(t *testing.T) {
	s := Set{}
	s.Add(1)

	slice := []int{2, 3, 4, 5}
	s.AddSlice(slice, func(i int) interface{} {
		return slice[i]
	})

	assert.Equal(t, 5, len(s))
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains(6))

	s.Union(NewSet(6, 7, 8, 9, 10))
	assert.Equal(t, 10, len(s))

	s.Remove(1)
	assert.Equal(t, 9, len(s))
}

package cast_test

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/cast"
)

func TestStringToIntSlice(t *testing.T) {
	cases := []struct {
		Input    string
		Expected []int
	}{
		{"1,2,3", []int{1, 2, 3}},
	}

	for _, c := range cases {
		r := cast.StringToIntSlice(c.Input, ",")
		assert.Equal(t, c.Expected, r)
	}
}

func TestStringToInt32Slice(t *testing.T) {
	cases := []struct {
		Input    string
		Expected []int32
	}{
		{"1,2,3", []int32{1, 2, 3}},
	}

	for _, c := range cases {
		r := cast.StringToInt32Slice(c.Input, ",")
		assert.Equal(t, c.Expected, r)
	}
}

func TestStringToInt64Slice(t *testing.T) {
	cases := []struct {
		Input    string
		Expected []int64
	}{
		{"1,2,3", []int64{1, 2, 3}},
	}

	for _, c := range cases {
		r := cast.StringToInt64Slice(c.Input, ",")
		assert.Equal(t, c.Expected, r)
	}
}

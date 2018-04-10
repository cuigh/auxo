package cast_test

import (
	"reflect"
	"testing"

	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/cast"
)

func TestTryToValue(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Type     reflect.Type
		Expected interface{}
	}{
		{"true", reflects.TypeBool, true},
		{"abc", reflects.TypeString, "abc"},
		{"1", reflects.TypeInt, int(1)},
		{"1", reflects.TypeInt8, int8(1)},
		{"1", reflects.TypeInt16, int16(1)},
		{"1", reflects.TypeInt32, int32(1)},
		{"1", reflects.TypeInt64, int64(1)},
	}

	for _, c := range cases {
		v, err := cast.TryToValue(c.Input, c.Type)
		assert.NoError(t, err)
		assert.Equal(t, c.Expected, v.Interface())
	}
}

func TestTryToSlice(t *testing.T) {
	cases := []struct {
		Input interface{}
		Type  reflect.Type
	}{
		{"", reflects.TypeString},
		{"1,false,1", reflects.TypeBool},
		{"1,2,3", reflects.TypeString},
		{"1,2,3", reflects.TypeInt},
		{"1,2,3", reflects.TypeInt8},
		{"1,2,3", reflects.TypeInt16},
		{"1,2,3", reflects.TypeInt32},
		{"1,2,3", reflects.TypeInt64},
	}

	for _, c := range cases {
		slice, err := cast.TryToSlice(c.Input, c.Type)
		assert.NoError(t, err)
		t.Logf("%#v", slice)
	}
}

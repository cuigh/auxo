package reflects_test

import (
	"reflect"
	"testing"

	"encoding/json"

	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/test/assert"
)

func TestSliceLen(t *testing.T) {
	arr := []int{1, 2, 3}
	l := reflects.SliceLen(arr)
	assert.Equal(t, len(arr), l)
}

func TestSliceInfo_GetString(t *testing.T) {
	slice := []string{"hello", "world"}
	ptr := reflects.SlicePtr(slice)
	si := reflects.NewSliceInfo(reflect.TypeOf(slice))
	assert.Equal(t, "hello", si.GetString(ptr, 0))
	assert.Equal(t, "world", si.GetString(ptr, 1))
}

func TestSliceInfo_SetString(t *testing.T) {
	slice := []string{"", ""}
	ptr := reflects.SlicePtr(slice)
	si := reflects.NewSliceInfo(reflect.TypeOf(slice))
	si.SetString(ptr, 0, "hello")
	si.SetString(ptr, 1, "world")
	assert.Equal(t, "hello", si.GetString(ptr, 0))
	assert.Equal(t, "world", si.GetString(ptr, 1))
}

func BenchmarkSlice1(b *testing.B) {
	slice := []string{"hello", "world"}
	v := reflect.ValueOf(slice)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v.Index(1).String()
	}
}

func BenchmarkSlice2(b *testing.B) {
	slice := []string{"hello", "world"}
	si := reflects.NewSliceInfo(reflect.TypeOf(slice))
	ptr := reflects.SlicePtr(slice)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		si.GetString(ptr, 1)
	}
}

func TestSliceValue(t *testing.T) {
	v := struct {
		Slice    []int
		SlicePtr *[]int
	}{}
	cases := []struct {
		Value reflect.Value
	}{
		{reflect.ValueOf(&v).Elem().FieldByName("Slice")},
		{reflect.ValueOf(&v).Elem().FieldByName("SlicePtr")},
	}

	for _, c := range cases {
		v := reflects.SliceOf(c.Value)

		v.Add(1)
		v.AddValue(reflect.ValueOf(2))
		v.Set(0, 3)
		v.SetValue(1, reflect.ValueOf(4))

		assertEqual(t, []int{3, 4}, c.Value.Interface())
	}
}

func TestArrayValue(t *testing.T) {
	v := struct {
		Array    [2]int
		ArrayPtr *[2]int
	}{}
	cases := []struct {
		Value reflect.Value
	}{
		{reflect.ValueOf(&v).Elem().FieldByName("Array")},
		{reflect.ValueOf(&v).Elem().FieldByName("ArrayPtr")},
	}

	for _, c := range cases {
		v := reflects.ArrayOf(c.Value)

		v.Set(0, 1)
		v.SetValue(1, reflect.ValueOf(2))

		assertEqual(t, []int{1, 2}, c.Value.Interface())
	}
}

func assertEqual(t *testing.T, x, y interface{}) {
	t.Helper()

	bx, err := json.Marshal(x)
	assert.NoError(t, err)

	by, err := json.Marshal(y)
	assert.NoError(t, err)

	assert.Equal(t, string(bx), string(by))
}

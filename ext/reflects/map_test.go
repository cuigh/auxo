package reflects_test

import (
	"reflect"
	"testing"

	"github.com/cuigh/auxo/ext/reflects"
)

func TestMapValue(t *testing.T) {
	v := struct {
		Map    map[int]string
		MapPtr *map[int]string
	}{}
	cases := []struct {
		Value  reflect.Value
		Key1   int
		Key2   int
		Value1 string
		Value2 string
	}{
		{reflect.ValueOf(&v).Elem().FieldByName("Map"), 1, 2, "x", "y"},
		{reflect.ValueOf(&v).Elem().FieldByName("MapPtr"), 1, 2, "x", "y"},
	}

	for _, c := range cases {
		v := reflects.MapOf(c.Value)

		v.Set(c.Key1, c.Value1)
		v.SetValue(reflect.ValueOf(c.Key2), reflect.ValueOf(c.Value2))

		t.Log(v.Interface())
	}
}

package reflects

import (
	"reflect"
)

type MapValue reflect.Value

func MapOf(v reflect.Value) MapValue {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	return MapValue(v)
}

func (mv MapValue) Set(key, value interface{}) {
	mv.SetValue(reflect.ValueOf(key), reflect.ValueOf(value))
}

func (mv MapValue) SetValue(key, value reflect.Value) {
	v := (reflect.Value)(mv)
	v.SetMapIndex(key, value)
}

func (mv MapValue) Delete(key interface{}) {
	mv.DeleteValue(reflect.ValueOf(key))
}

func (mv MapValue) DeleteValue(key reflect.Value) {
	v := (reflect.Value)(mv)
	v.SetMapIndex(key, reflect.Zero(v.Type().Elem()))
}

func (mv MapValue) Get(key interface{}) interface{} {
	return mv.GetValue(reflect.ValueOf(key)).Interface()
}

func (mv MapValue) GetValue(key reflect.Value) reflect.Value {
	v := (reflect.Value)(mv)
	return v.MapIndex(key)
}

func (mv MapValue) Interface() interface{} {
	v := (reflect.Value)(mv)
	return v.Interface()
}

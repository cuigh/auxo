package reflects

import (
	"reflect"
	"unsafe"

	"github.com/cuigh/auxo/errors"
)

//var (
//	sliceInvalidGetter = func(ptr uintptr, t reflect.Type, i int) interface{} {
//		panic(errors.Format("unsupported element type: %v", t))
//	}
//	sliceInvalidSetter = func(ptr uintptr, t reflect.Type, i int, v interface{}) {
//		panic(errors.Format("unsupported element type: %v", t))
//	}
//)

func SlicePtr(i interface{}) uintptr {
	ptr := Pointer(i)
	header := (*reflect.SliceHeader)(ptr)
	return header.Data
}

func SliceLen(i interface{}) int {
	ptr := Pointer(i)
	header := (*reflect.SliceHeader)(ptr)
	return header.Len
}

func SlicePointer(i interface{}) unsafe.Pointer {
	return unsafe.Pointer(SlicePtr(i))
}

type SliceInfo struct {
	elemType reflect.Type
}

func NewSliceInfo(t reflect.Type) *SliceInfo {
	return &SliceInfo{
		elemType: t.Elem(),
	}
}

func (s *SliceInfo) Get(ptr uintptr, i int) interface{} {
	switch s.elemType.Kind() {
	case reflect.String:
		return s.GetString(ptr, i)
	case reflect.Bool:
		return s.GetBool(ptr, i)
	case reflect.Int:
		return s.GetInt(ptr, i)
	case reflect.Int8:
		return s.GetInt8(ptr, i)
	case reflect.Int16:
		return s.GetInt16(ptr, i)
	case reflect.Int32:
		return s.GetInt32(ptr, i)
	case reflect.Int64:
		return s.GetInt64(ptr, i)
	case reflect.Uint:
		return s.GetUint(ptr, i)
	case reflect.Uint8:
		return s.GetUint8(ptr, i)
	case reflect.Uint16:
		return s.GetUint16(ptr, i)
	case reflect.Uint32:
		return s.GetUint32(ptr, i)
	case reflect.Uint64:
		return s.GetUint64(ptr, i)
	default:
		panic(errors.Format("unsupported element type: %v", s.elemType))
	}
}

func (s *SliceInfo) Set(ptr uintptr, i int, v interface{}) {
	switch s.elemType.Kind() {
	case reflect.String:
		s.SetString(ptr, i, v.(string))
	case reflect.Bool:
		s.SetBool(ptr, i, v.(bool))
	case reflect.Int:
		s.SetInt(ptr, i, v.(int))
	case reflect.Int8:
		s.SetInt8(ptr, i, v.(int8))
	case reflect.Int16:
		s.SetInt16(ptr, i, v.(int16))
	case reflect.Int32:
		s.SetInt32(ptr, i, v.(int32))
	case reflect.Int64:
		s.SetInt64(ptr, i, v.(int64))
	case reflect.Uint:
		s.SetUint(ptr, i, v.(uint))
	case reflect.Uint8:
		s.SetUint8(ptr, i, v.(uint8))
	case reflect.Uint16:
		s.SetUint16(ptr, i, v.(uint16))
	case reflect.Uint32:
		s.SetUint32(ptr, i, v.(uint32))
	case reflect.Uint64:
		s.SetUint64(ptr, i, v.(uint64))
	default:
		panic(errors.Format("unsupported element type: %v", s.elemType))
	}
}

func (s *SliceInfo) GetString(ptr uintptr, i int) string {
	return *((*string)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetString(ptr uintptr, i int, v string) {
	*((*string)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetBool(ptr uintptr, i int) bool {
	return *((*bool)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetBool(ptr uintptr, i int, v bool) {
	*((*bool)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetInt(ptr uintptr, i int) int {
	return *((*int)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetInt(ptr uintptr, i int, v int) {
	*((*int)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetInt8(ptr uintptr, i int) int8 {
	return *((*int8)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetInt8(ptr uintptr, i int, v int8) {
	*((*int8)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetInt16(ptr uintptr, i int) int16 {
	return *((*int16)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetInt16(ptr uintptr, i int, v int16) {
	*((*int16)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetInt32(ptr uintptr, i int) int32 {
	return *((*int32)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetInt32(ptr uintptr, i int, v int32) {
	*((*int32)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetInt64(ptr uintptr, i int) int64 {
	return *((*int64)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetInt64(ptr uintptr, i int, v int64) {
	*((*int64)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetUint(ptr uintptr, i int) uint {
	return *((*uint)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetUint(ptr uintptr, i int, v uint) {
	*((*uint)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetUint8(ptr uintptr, i int) uint8 {
	return *((*uint8)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetUint8(ptr uintptr, i int, v uint8) {
	*((*uint8)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetUint16(ptr uintptr, i int) uint16 {
	return *((*uint16)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetUint16(ptr uintptr, i int, v uint16) {
	*((*uint16)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetUint32(ptr uintptr, i int) uint32 {
	return *((*uint32)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetUint32(ptr uintptr, i int, v uint32) {
	*((*uint32)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

func (s *SliceInfo) GetUint64(ptr uintptr, i int) uint64 {
	return *((*uint64)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i))))
}

func (s *SliceInfo) SetUint64(ptr uintptr, i int, v uint64) {
	*((*uint64)(unsafe.Pointer(ptr + s.elemType.Size()*uintptr(i)))) = v
}

type SliceValue reflect.Value

func SliceOf(v reflect.Value) SliceValue {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}
	return SliceValue(v)
}

func (sv SliceValue) Set(i int, x interface{}) {
	sv.SetValue(i, reflect.ValueOf(x))
}

func (sv SliceValue) SetValue(i int, x reflect.Value) {
	v := (reflect.Value)(sv)
	v.Index(i).Set(x)
}

func (sv SliceValue) Add(x interface{}) {
	sv.AddValue(reflect.ValueOf(x))
}

func (sv SliceValue) AddValue(x ...reflect.Value) {
	v := (reflect.Value)(sv)
	v.Set(reflect.Append(v, x...))
}

func (sv SliceValue) Get(i int) interface{} {
	return sv.GetValue(i).Interface()
}

func (sv SliceValue) GetValue(i int) reflect.Value {
	v := (reflect.Value)(sv)
	return v.Index(i)
}

func (sv SliceValue) Interface() interface{} {
	v := (reflect.Value)(sv)
	return v.Interface()
}

type ArrayValue reflect.Value

func ArrayOf(v reflect.Value) ArrayValue {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}
	return ArrayValue(v)
}

func (av ArrayValue) Set(i int, x interface{}) {
	av.SetValue(i, reflect.ValueOf(x))
}

func (av ArrayValue) SetValue(i int, x reflect.Value) {
	v := (reflect.Value)(av)
	v.Index(i).Set(x)
}

func (av ArrayValue) Get(i int) interface{} {
	return av.GetValue(i).Interface()
}

func (av ArrayValue) GetValue(i int) reflect.Value {
	v := (reflect.Value)(av)
	return v.Index(i)
}

func (av ArrayValue) Interface() interface{} {
	v := (reflect.Value)(av)
	return v.Interface()
}

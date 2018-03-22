package reflects

import (
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/cuigh/auxo/errors"
)

var (
	defaultFieldAccessor = &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			panic(errors.Format("unsupported field: %v(%v)", f.Name, f.Type))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			panic(errors.Format("unsupported field: %v(%v)", f.Name, f.Type))
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			panic(errors.Format("unsupported field: %v(%v)", f.Name, f.Type))
		},
	}
)

//type StructInfo struct {
//	FieldPointers map[string]unsafe.Pointer
//}
//
//func NewStructInfo(t reflect.Type) *StructInfo {
//	if t.Kind() == reflect.Ptr {
//		t = t.Elem()
//	}
//	if t.Kind() != reflect.Struct {
//		panic("reflects: NewStructInfo of non-struct type")
//	}
//
//	si := &StructInfo{}
//	return si
//}

var fieldAccessors = map[reflect.Type]*FieldAccessor{}

type FieldAccessor struct {
	Get        func(ptr uintptr, f *reflect.StructField) interface{}
	Set        func(ptr uintptr, f *reflect.StructField, v interface{})
	GetPointer func(ptr uintptr, f *reflect.StructField) interface{}
}

func RegisterFieldAccessor(t reflect.Type, accessor *FieldAccessor) {
	fieldAccessors[t] = accessor
}

//type FieldGetter func(ptr uintptr, f *reflect.StructField) interface{}
//type FieldSetter func(ptr uintptr, f *reflect.StructField, v interface{})

type FieldInfo struct {
	*reflect.StructField
	accessor *FieldAccessor
}

func NewFieldInfo(f *reflect.StructField) *FieldInfo {
	fi := &FieldInfo{
		StructField: f,
		accessor:    fieldAccessors[f.Type],
	}
	if fi.accessor == nil {
		switch fi.Type.Kind() {
		case reflect.Int:
			fi.accessor = fieldAccessors[TypeInt]
		case reflect.Int8:
			fi.accessor = fieldAccessors[TypeInt8]
		case reflect.Int16:
			fi.accessor = fieldAccessors[TypeInt16]
		case reflect.Int32:
			fi.accessor = fieldAccessors[TypeInt32]
		case reflect.Int64:
			fi.accessor = fieldAccessors[TypeInt64]
		case reflect.Uint:
			fi.accessor = fieldAccessors[TypeUint]
		case reflect.Uint8:
			fi.accessor = fieldAccessors[TypeUint8]
		case reflect.Uint16:
			fi.accessor = fieldAccessors[TypeUint16]
		case reflect.Uint32:
			fi.accessor = fieldAccessors[TypeUint32]
		case reflect.Uint64:
			fi.accessor = fieldAccessors[TypeUint64]
		case reflect.Float32:
			fi.accessor = fieldAccessors[TypeFloat32]
		case reflect.Float64:
			fi.accessor = fieldAccessors[TypeFloat64]
		case reflect.Bool:
			fi.accessor = fieldAccessors[TypeBool]
		case reflect.String:
			fi.accessor = fieldAccessors[TypeString]
		default:
			fi.accessor = defaultFieldAccessor
		}
	}
	return fi
}

func (f *FieldInfo) Get(ptr uintptr) interface{} {
	return f.accessor.Get(ptr, f.StructField)
}

func (f *FieldInfo) Set(ptr uintptr, v interface{}) {
	f.accessor.Set(ptr, f.StructField, v)
}

func (f *FieldInfo) GetPointer(ptr uintptr) interface{} {
	return f.accessor.GetPointer(ptr, f.StructField)
}

type StructTag reflect.StructTag

func (tag StructTag) All() map[string]string {
	m := make(map[string]string)
	// see: std > reflect.StructTag.Lookup
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err == nil {
			m[name] = value
		}
		// todo: log or print error?
	}
	return m
}

func (tag StructTag) Find(key string, alias ...string) string {
	t := reflect.StructTag(tag)
	if v, ok := t.Lookup(key); ok {
		return v
	}
	for _, k := range alias {
		if v, ok := t.Lookup(k); ok {
			return v
		}
	}
	return ""
}

type StructValue reflect.Value

func StructOf(v reflect.Value) StructValue {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}
	return StructValue(v)
}

func (sv StructValue) VisitFields(fn func(fv reflect.Value, fi *reflect.StructField) error) error {
	v := (reflect.Value)(sv)
	t := v.Type()
	for i, num := 0, t.NumField(); i < num; i++ {
		sf := t.Field(i)
		if err := fn(v.Field(i), &sf); err != nil {
			return err
		}
	}
	return nil
}

func (sv StructValue) VisitMethods(fn func(mv reflect.Value, mi *reflect.Method) error) error {
	v := (reflect.Value)(sv)
	t := v.Type()
	for i, num := 0, t.NumMethod(); i < num; i++ {
		mi := t.Method(i)
		if err := fn(v.Method(i), &mi); err != nil {
			return err
		}
	}
	return nil
}

func (sv StructValue) Interface() interface{} {
	v := (reflect.Value)(sv)
	return v.Interface()
}

func init() {
	RegisterFieldAccessor(TypeBool, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*bool)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*bool)(unsafe.Pointer(ptr + f.Offset))) = v.(bool)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*bool)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeString, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*string)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*string)(unsafe.Pointer(ptr + f.Offset))) = v.(string)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*string)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeTime, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*time.Time)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*time.Time)(unsafe.Pointer(ptr + f.Offset))) = v.(time.Time)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*time.Time)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeInt, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*int)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*int)(unsafe.Pointer(ptr + f.Offset))) = v.(int)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*int)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeInt8, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*int8)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*int8)(unsafe.Pointer(ptr + f.Offset))) = v.(int8)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*int8)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeInt16, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*int16)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*int16)(unsafe.Pointer(ptr + f.Offset))) = v.(int16)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*int16)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeInt32, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*int32)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*int32)(unsafe.Pointer(ptr + f.Offset))) = v.(int32)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*int32)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeInt64, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*int64)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*int64)(unsafe.Pointer(ptr + f.Offset))) = v.(int64)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*int64)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeUint, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*uint)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*uint)(unsafe.Pointer(ptr + f.Offset))) = v.(uint)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*uint)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeUint8, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*uint8)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*uint8)(unsafe.Pointer(ptr + f.Offset))) = v.(uint8)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*uint8)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeUint16, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*uint16)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*uint16)(unsafe.Pointer(ptr + f.Offset))) = v.(uint16)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*uint16)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeUint32, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*uint32)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*uint32)(unsafe.Pointer(ptr + f.Offset))) = v.(uint32)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*uint32)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeUint64, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*uint64)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*uint64)(unsafe.Pointer(ptr + f.Offset))) = v.(uint64)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*uint64)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeFloat32, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*float32)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*float32)(unsafe.Pointer(ptr + f.Offset))) = v.(float32)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*float32)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeFloat64, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*float64)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*float64)(unsafe.Pointer(ptr + f.Offset))) = v.(float64)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*float64)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	RegisterFieldAccessor(TypeBytes, &FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*[]byte)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*[]byte)(unsafe.Pointer(ptr + f.Offset))) = v.([]byte)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*[]byte)(unsafe.Pointer(ptr + f.Offset))
		},
	})
}

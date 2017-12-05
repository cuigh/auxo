package reflects

import (
	"context"
	"reflect"
	"time"
	"unsafe"

	"github.com/cuigh/auxo/errors"
)

var (
	TypeError    = reflect.TypeOf((*error)(nil)).Elem()
	TypeContext  = reflect.TypeOf((*context.Context)(nil)).Elem()
	TypeString   = reflect.TypeOf("")
	TypeBool     = reflect.TypeOf(true)
	TypeBytes    = reflect.TypeOf([]byte{})
	TypeInt      = reflect.TypeOf(int(0))
	TypeInt8     = reflect.TypeOf(int8(0))
	TypeInt16    = reflect.TypeOf(int16(0))
	TypeInt32    = reflect.TypeOf(int32(0))
	TypeInt64    = reflect.TypeOf(int64(0))
	TypeUint     = reflect.TypeOf(uint(0))
	TypeUint8    = reflect.TypeOf(uint8(0))
	TypeUint16   = reflect.TypeOf(uint16(0))
	TypeUint32   = reflect.TypeOf(uint32(0))
	TypeUint64   = reflect.TypeOf(uint64(0))
	TypeFloat32  = reflect.TypeOf(float32(0))
	TypeFloat64  = reflect.TypeOf(float64(0))
	TypeTime     = reflect.TypeOf(time.Time{})
	TypeDuration = reflect.TypeOf(time.Duration(0))
	//TypeIntPtr   = reflect.PtrTo(TypeInt)
)

var (
	ZeroError = reflect.Zero(TypeError)
)

func Error(err error) reflect.Value {
	if err == nil {
		return ZeroError
	} else {
		return reflect.ValueOf(err).Convert(TypeError)
	}
}

func Value(i interface{}, t reflect.Type) reflect.Value {
	if i == nil {
		return reflect.Zero(t)
	} else {
		return reflect.ValueOf(i)
	}
}

// Interface convert v to it's interface value, if v is nil value, return nil directly
func Interface(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
		if v.IsNil() {
			return nil
		}
	}
	return v.Interface()
}

func Indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func IsEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		v = Indirect(v)
		return IsEmpty(v)
	}
	return false
}

func Pointer(i interface{}) unsafe.Pointer {
	return (*eface)(unsafe.Pointer(&i)).data
}

func Ptr(i interface{}) uintptr {
	return uintptr(Pointer(i))
}

// eface is empty interface
type eface struct {
	dt   *struct{}
	data unsafe.Pointer
}

// Connect fill all func fields of dst with methods of src.
func Connect(dst, src interface{}) error {
	vd := reflect.ValueOf(dst).Elem()
	td := vd.Type()
	vs := reflect.ValueOf(src)
	ts := vs.Type()
	for i := 0; i < td.NumField(); i++ {
		sf := td.Field(i)
		if sf.Type.Kind() != reflect.Func {
			continue
		}

		m := vs.MethodByName(sf.Name)
		if !m.IsValid() {
			return errors.Format("type [%s] doesn't contain method [%s]", ts.Name(), sf.Name)
		}

		f := vd.Field(i)
		if !m.Type().ConvertibleTo(f.Type()) {
			return errors.Format("%s: %v is not assignable to %v", sf.Name, m.Type(), f.Type())
		}
		f.Set(m)
	}
	return nil
}

//func UnmarshalValue(v reflect.Value,
//	namer func(sf *reflect.StructField) string,
//	getter func(name string) interface{},
//	setter func(f reflect.Value, value interface{}) (bool, error)) (err error) {
//	if v.Kind() == reflect.Ptr {
//		v = v.Elem()
//	}
//	t := v.Type()
//	count := t.NumField()
//	for i := 0; i < count; i++ {
//		sf := t.Field(i)
//		err = setField(v.Field(i), &sf, namer, getter, setter)
//		if err != nil {
//			return
//		}
//	}
//	return
//}
//
//func setField(f reflect.Value, sf *reflect.StructField, namer func(sf *reflect.StructField) string, getter func(name string) interface{}, setter func(f reflect.Value, value interface{}) (bool, error)) error {
//	name := namer(sf)
//	if name == "" {
//		name = texts.Rename(sf.Name, texts.Lower)
//	}
//
//	opt := getter(name)
//	if opt == nil {
//		return nil
//	}
//
//	if f.Kind() == reflect.Ptr {
//		if f.IsNil() {
//			f.Set(reflect.New(f.Type().Elem()))
//		}
//		f = f.Elem()
//	}
//
//	if setter != nil {
//		if ok, err := setter(f, opt); err != nil {
//			return err
//		} else if ok {
//			return nil
//		}
//	}
//
//	if sf.Type == TypeDuration {
//		f.Set(reflect.ValueOf(cast.ToDuration(opt)))
//	} else if sf.Type == TypeTime {
//		f.Set(reflect.ValueOf(cast.ToTime(opt)))
//	} else {
//		switch f.Kind() {
//		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//			f.SetInt(cast.ToInt64(opt))
//		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
//			f.SetUint(uint64(cast.ToInt64(opt)))
//		case reflect.Float32, reflect.Float64:
//			f.SetFloat(cast.ToFloat64(opt))
//		case reflect.String:
//			f.SetString(cast.ToString(opt))
//		case reflect.Bool:
//			f.SetBool(cast.ToBool(opt))
//		case reflect.Struct:
//			return UnmarshalValue(f, namer, getter, setter)
//		case reflect.Slice:
//			slice, err := cast.TryToSliceValue(opt, f.Type().Elem())
//			if err != nil {
//				return err
//			}
//			f.Set(slice)
//		default:
//			return errors.Format("can't decode option '%s' to type %v", name, sf.Type)
//		}
//	}
//}

type SimpleValue reflect.Value

func SimpleOf(v reflect.Value) SimpleValue {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}
	return SimpleValue(v)
}

func (sv SimpleValue) Set(x interface{}) {
	sv.SetValue(reflect.ValueOf(x))
}

func (sv SimpleValue) SetValue(x reflect.Value) {
	(reflect.Value)(sv).Set(x)
}

func (sv SimpleValue) Interface() interface{} {
	v := (reflect.Value)(sv)
	return v.Interface()
}

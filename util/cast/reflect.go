package cast

import (
	"reflect"
	"strings"

	"github.com/cuigh/auxo/ext/reflects"
)

func TryToValue(i interface{}, t reflect.Type) (v reflect.Value, err error) {
	var value interface{}

	if t == reflects.TypeDuration {
		value, err = TryToDuration(i)
	} else if t == reflects.TypeTime {
		value, err = TryToTime(i)
	} else {
		switch t.Kind() {
		case reflect.Bool:
			value, err = TryToBool(i)
		case reflect.String:
			value = ToString(i)
		case reflect.Int:
			value, err = TryToInt(i)
		case reflect.Int8:
			value, err = TryToInt8(i)
		case reflect.Int16:
			value, err = TryToInt16(i)
		case reflect.Int32:
			value, err = TryToInt32(i)
		case reflect.Int64:
			value, err = TryToInt64(i)
		case reflect.Uint:
			value, err = TryToUint(i)
		case reflect.Uint8:
			value, err = TryToUint8(i)
		case reflect.Uint16:
			value, err = TryToUint16(i)
		case reflect.Uint32:
			value, err = TryToUint32(i)
		case reflect.Uint64:
			value, err = TryToUint64(i)
		case reflect.Float32:
			value, err = TryToFloat32(i)
		case reflect.Float64:
			value, err = TryToFloat64(i)
		case reflect.Slice:
			return TryToSliceValue(i, t.Elem())
		default:
			err = castError(i, t.String())
		}
	}
	if err == nil {
		v = reflect.ValueOf(value).Convert(t)
	}
	return
}

// TryToSliceValue cast interface value to a slice.
// Argument t is element type of slice.
func TryToSliceValue(i interface{}, t reflect.Type) (slice reflect.Value, err error) {
	if s, ok := i.(string); ok {
		if s == "" {
			i = []string(nil)
		} else {
			i = strings.Split(s, ",")
		}
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		err = castError(i, "[]"+t.String())
		return
	}

	if t == v.Type().Elem() {
		return v, nil
	}

	length := v.Len()
	slice = reflect.MakeSlice(reflect.SliceOf(t), length, length)
	for k := 0; k < length; k++ {
		var value reflect.Value
		value, err = TryToValue(v.Index(k).Interface(), t)
		if err != nil {
			return
		}
		slice.Index(k).Set(value)
	}
	return slice, nil
}

func TryToSlice(i interface{}, t reflect.Type) (interface{}, error) {
	v, err := TryToSliceValue(i, t)
	if err != nil {
		return nil, err
	}
	return v.Interface(), nil
}

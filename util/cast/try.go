package cast

import (
	"fmt"
	"strconv"
	"time"
)

// TryToBool casts an empty interface to a bool.
func TryToBool(i interface{}) (b bool, err error) {
	switch v := i.(type) {
	case nil:
	case bool:
		b = v
	case *bool:
		b = *v
	case string:
		b, err = strconv.ParseBool(v)
	case *string:
		b, err = strconv.ParseBool(*v)
	case int, int8, int16, int32, int64:
		b = v != 0
	default:
		err = castError(i, "bool")
	}
	return
}

// TryToFloat32 casts an empty interface to a float32.
func TryToFloat32(i interface{}) (r float32, err error) {
	switch t := i.(type) {
	case float32:
		r = t
	case *float32:
		r = *t
	case float64:
		r = float32(t)
	case int:
		r = float32(t)
	case int8:
		r = float32(t)
	case int16:
		r = float32(t)
	case int32:
		r = float32(t)
	case int64:
		r = float32(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v float64
		if v, err = strconv.ParseFloat(t, 32); err == nil {
			r = float32(v)
		}
	default:
		err = castError(i, "float32")
	}
	return
}

// TryToFloat64 casts an empty interface to a float64.
func TryToFloat64(i interface{}) (r float64, err error) {
	switch t := i.(type) {
	case float32:
		r = float64(t)
	case float64:
		r = t
	case *float64:
		r = *t
	case int:
		r = float64(t)
	case int8:
		r = float64(t)
	case int16:
		r = float64(t)
	case int32:
		r = float64(t)
	case int64:
		r = float64(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		r, err = strconv.ParseFloat(t, 64)
	default:
		err = castError(i, "float64")
	}
	return
}

// TryToInt casts an empty interface to an int.
func TryToInt(i interface{}) (r int, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case int:
		r = t
	case *int:
		r = *t
	case int8:
		r = int(t)
	case int16:
		r = int(t)
	case int32:
		r = int(t)
	case int64:
		r = int(t)
	case uint:
		r = int(t)
	case uint8:
		r = int(t)
	case uint16:
		r = int(t)
	case uint32:
		r = int(t)
	case uint64:
		r = int(t)
	case float32:
		r = int(t)
	case float64:
		r = int(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		r, err = strconv.Atoi(t)
	default:
		err = castError(i, "int")
	}
	return
}

// TryToInt8 casts an empty interface to an int8.
func TryToInt8(i interface{}) (r int8, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case int:
		r = int8(t)
	case int8:
		r = t
	case *int8:
		r = *t
	case int16:
		r = int8(t)
	case int32:
		r = int8(t)
	case int64:
		r = int8(t)
	case uint:
		r = int8(t)
	case uint8:
		r = int8(t)
	case uint16:
		r = int8(t)
	case uint32:
		r = int8(t)
	case uint64:
		r = int8(t)
	case float32:
		r = int8(t)
	case float64:
		r = int8(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v int64
		if v, err = strconv.ParseInt(t, 10, 8); err == nil {
			r = int8(v)
		}
	default:
		err = castError(i, "int8")
	}
	return
}

// TryToInt16 casts an empty interface to an int16.
func TryToInt16(i interface{}) (r int16, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case int:
		r = int16(t)
	case int8:
		r = int16(t)
	case int16:
		r = t
	case *int16:
		r = *t
	case int32:
		r = int16(t)
	case int64:
		r = int16(t)
	case uint:
		r = int16(t)
	case uint8:
		r = int16(t)
	case uint16:
		r = int16(t)
	case uint32:
		r = int16(t)
	case uint64:
		r = int16(t)
	case float32:
		r = int16(t)
	case float64:
		r = int16(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v int64
		if v, err = strconv.ParseInt(t, 10, 16); err == nil {
			r = int16(v)
		}
	default:
		err = castError(i, "int16")
	}
	return
}

// TryToInt32 casts an empty interface to an int32.
func TryToInt32(i interface{}) (r int32, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case int:
		r = int32(t)
	case int8:
		r = int32(t)
	case int16:
		r = int32(t)
	case int32:
		r = t
	case *int32:
		r = *t
	case int64:
		r = int32(t)
	case uint:
		r = int32(t)
	case uint8:
		r = int32(t)
	case uint16:
		r = int32(t)
	case uint32:
		r = int32(t)
	case uint64:
		r = int32(t)
	case float32:
		r = int32(t)
	case float64:
		r = int32(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v int64
		if v, err = strconv.ParseInt(t, 10, 32); err == nil {
			r = int32(v)
		}
	default:
		err = castError(i, "int32")
	}
	return
}

// TryToInt64 casts an empty interface to an int64.
func TryToInt64(i interface{}) (r int64, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case int:
		r = int64(t)
	case int8:
		r = int64(t)
	case int16:
		r = int64(t)
	case int32:
		r = int64(t)
	case int64:
		r = t
	case *int64:
		r = *t
	case uint:
		r = int64(t)
	case uint8:
		r = int64(t)
	case uint16:
		r = int64(t)
	case uint32:
		r = int64(t)
	case uint64:
		r = int64(t)
	case float32:
		r = int64(t)
	case float64:
		r = int64(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		r, err = strconv.ParseInt(t, 10, 64)
	default:
		err = castError(i, "int64")
	}
	return
}

// TryToUint casts an empty interface to a uint.
func TryToUint(i interface{}) (r uint, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case uint:
		r = t
	case *uint:
		r = *t
	case uint8:
		r = uint(t)
	case uint16:
		r = uint(t)
	case uint32:
		r = uint(t)
	case uint64:
		r = uint(t)
	case int:
		r = uint(t)
	case int8:
		r = uint(t)
	case int16:
		r = uint(t)
	case int32:
		r = uint(t)
	case int64:
		r = uint(t)
	case float32:
		r = uint(t)
	case float64:
		r = uint(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v uint64
		if v, err = strconv.ParseUint(t, 10, 64); err == nil {
			r = uint(v)
		}
	default:
		err = castError(i, "uint")
	}
	return
}

// TryToUint8 casts an empty interface to an uint8.
func TryToUint8(i interface{}) (r uint8, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case uint:
		r = uint8(t)
	case uint8:
		r = t
	case *uint8:
		r = *t
	case uint16:
		r = uint8(t)
	case uint32:
		r = uint8(t)
	case uint64:
		r = uint8(t)
	case int:
		r = uint8(t)
	case int8:
		r = uint8(t)
	case int16:
		r = uint8(t)
	case int32:
		r = uint8(t)
	case int64:
		r = uint8(t)
	case float32:
		r = uint8(t)
	case float64:
		r = uint8(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v uint64
		if v, err = strconv.ParseUint(t, 10, 8); err == nil {
			r = uint8(v)
		}
	default:
		err = castError(i, "uint8")
	}
	return
}

// TryToUint16 casts an empty interface to a uint16.
func TryToUint16(i interface{}) (r uint16, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case uint:
		r = uint16(t)
	case uint8:
		r = uint16(t)
	case uint16:
		r = t
	case *uint16:
		r = *t
	case uint32:
		r = uint16(t)
	case uint64:
		r = uint16(t)
	case int:
		r = uint16(t)
	case int8:
		r = uint16(t)
	case int16:
		r = uint16(t)
	case int32:
		r = uint16(t)
	case int64:
		r = uint16(t)
	case float32:
		r = uint16(t)
	case float64:
		r = uint16(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v uint64
		if v, err = strconv.ParseUint(t, 10, 16); err == nil {
			r = uint16(v)
		}
	default:
		err = castError(i, "uint16")
	}
	return
}

// TryToUint32 casts an empty interface to a uint32.
func TryToUint32(i interface{}) (r uint32, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case uint:
		r = uint32(t)
	case uint8:
		r = uint32(t)
	case uint16:
		r = uint32(t)
	case uint32:
		r = t
	case *uint32:
		r = *t
	case uint64:
		r = uint32(t)
	case int:
		r = uint32(t)
	case int8:
		r = uint32(t)
	case int16:
		r = uint32(t)
	case int32:
		r = uint32(t)
	case int64:
		r = uint32(t)
	case float32:
		r = uint32(t)
	case float64:
		r = uint32(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		var v uint64
		if v, err = strconv.ParseUint(t, 10, 32); err == nil {
			r = uint32(v)
		}
	default:
		err = castError(i, "uint32")
	}
	return
}

// TryToUint64 casts an empty interface to a uint64.
func TryToUint64(i interface{}) (r uint64, err error) {
	switch t := i.(type) {
	case nil:
		r = 0
	case uint:
		r = uint64(t)
	case uint8:
		r = uint64(t)
	case uint16:
		r = uint64(t)
	case uint32:
		r = uint64(t)
	case uint64:
		r = uint64(t)
	case *uint64:
		r = *t
	case int:
		r = uint64(t)
	case int8:
		r = uint64(t)
	case int16:
		r = uint64(t)
	case int32:
		r = uint64(t)
	case int64:
		r = uint64(t)
	case float32:
		r = uint64(t)
	case float64:
		r = uint64(t)
	case bool:
		if t {
			r = 1
		}
	case string:
		r, err = strconv.ParseUint(t, 10, 64)
	default:
		err = castError(i, "uint64")
	}
	return
}

// TryToTime casts an empty interface to time.Time.
func TryToTime(i interface{}) (t time.Time, err error) {
	switch v := i.(type) {
	case time.Time:
		t = v
	case *time.Time:
		t = *v
	case string:
		t, err = StringToTime(v)
	case *string:
		t, err = StringToTime(*v)
	default:
		err = castError(i, "time.Time")
	}
	return
}

// TryToDuration casts an empty interface to time.Duration.
func TryToDuration(i interface{}) (d time.Duration, err error) {
	switch v := i.(type) {
	case time.Duration:
		d = v
	case *time.Duration:
		d = *v
	case int64:
		d = time.Duration(v)
	case *int64:
		d = time.Duration(*v)
	case string:
		d, err = time.ParseDuration(v)
	case *string:
		d, err = time.ParseDuration(*v)
	default:
		err = castError(i, "time.Duration")
	}
	return
}

// TryToInt32Slice casts an empty interface to []int32.
func TryToInt32Slice(i interface{}) (r []int32, err error) {
	switch v := i.(type) {
	case []int:
		r = make([]int32, len(v))
		for index, value := range v {
			r[index] = int32(value)
		}
	case []int32:
		r = v
	case *[]int32:
		r = *v
	case []int64:
		r = make([]int32, len(v))
		for index, value := range v {
			r[index] = int32(value)
		}
	case []string:
		r = make([]int32, len(v))
		for index, value := range v {
			r[index], err = TryToInt32(value)
			if err != nil {
				return
			}
		}
	default:
		err = castError(i, "[]int32")
	}
	return
}

// TryToInt64Slice casts an empty interface to []int64.
func TryToInt64Slice(i interface{}) (r []int64, err error) {
	switch v := i.(type) {
	case []int:
		r = make([]int64, len(v))
		for index, value := range v {
			r[index] = int64(value)
		}
	case []int32:
		r = make([]int64, len(v))
		for index, value := range v {
			r[index] = int64(value)
		}
	case []int64:
		r = v
	case *[]int64:
		r = *v
	case []string:
		r = make([]int64, len(v))
		for index, value := range v {
			r[index], err = TryToInt64(value)
			if err != nil {
				return
			}
		}
	default:
		err = castError(i, "[]int64")
	}
	return
}

func castError(i interface{}, t string) error {
	return fmt.Errorf("unable to cast %#v to %s", i, t)
}

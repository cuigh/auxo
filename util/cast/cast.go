package cast

import (
	"fmt"
	"html/template"
	"strconv"
	"time"
)

func ToBool(i interface{}, d ...bool) bool {
	v, err := TryToBool(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToFloat32(i interface{}, d ...float32) float32 {
	v, err := TryToFloat32(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToFloat64(i interface{}, d ...float64) float64 {
	v, err := TryToFloat64(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToInt(i interface{}, d ...int) int {
	v, err := TryToInt(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToInt8(i interface{}, d ...int8) int8 {
	v, err := TryToInt8(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToInt16(i interface{}, d ...int16) int16 {
	v, err := TryToInt16(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToInt32(i interface{}, d ...int32) int32 {
	v, err := TryToInt32(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToInt64(i interface{}, d ...int64) int64 {
	v, err := TryToInt64(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToUint(i interface{}, d ...uint) uint {
	v, err := TryToUint(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToUint8(i interface{}, d ...uint8) uint8 {
	v, err := TryToUint8(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToUint16(i interface{}, d ...uint16) uint16 {
	v, err := TryToUint16(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToUint32(i interface{}, d ...uint32) uint32 {
	v, err := TryToUint32(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToUint64(i interface{}, d ...uint64) uint64 {
	v, err := TryToUint64(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

// ToString casts an empty interface to a string.
func ToString(i interface{}) string {
	switch t := i.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case string:
		return t
	case *string:
		return *t
	case bool:
		return strconv.FormatBool(t)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case template.HTML:
		return string(t)
	case template.URL:
		return string(t)
	case fmt.Stringer:
		return t.String()
	case error:
		return t.Error()
	default:
		return fmt.Sprint(i)
	}
}

func ToTime(i interface{}, d ...time.Time) time.Time {
	v, err := TryToTime(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

func ToDuration(i interface{}, d ...time.Duration) time.Duration {
	v, err := TryToDuration(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

// ToInt32Slice casts an empty interface to []int32.
func ToInt32Slice(i interface{}, d ...[]int32) (r []int32) {
	v, err := TryToInt32Slice(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

// ToInt64Slice casts an empty interface to []int64.
func ToInt64Slice(i interface{}, d ...[]int64) (r []int64) {
	v, err := TryToInt64Slice(i)
	if err != nil && len(d) > 0 {
		return d[0]
	}
	return v
}

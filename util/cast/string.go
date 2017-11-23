package cast

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cuigh/auxo/ext/reflects"
)

var layouts = [...]string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	time.RFC3339,
	"2006-01-02T15:04:05", // iso8601 without timezone
	time.RFC1123Z,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	"2006-01-02 15:04:05Z07:00",
	"02 Jan 06 15:04 MST",
	"02 Jan 2006",
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05 -0700",
}

// StringToTime casts an empty interface to a time.Time.
func StringToTime(s string) (t time.Time, e error) {
	for _, dateType := range layouts {
		if t, e = time.Parse(dateType, s); e == nil {
			return
		}
	}
	return t, fmt.Errorf("unable to parse date: %s", s)
}

// TODO: refactor
func StringToType(s string, p reflect.Type) interface{} {
	switch p {
	case reflects.TypeTime:
		return ToTime(s)
	case reflects.TypeDuration:
		return ToDuration(s)
	default:
		switch p.Kind() {
		case reflect.String:
			return s
		case reflect.Bool:
			return ToBool(s)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return ToInt(s)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return ToUint(s)
		case reflect.Slice:
			et := p.Elem()
			if et.Kind() == reflect.String {
				return strings.Split(s, ",")
			}
		default:
			panic(errors.New("unsupported type: " + p.Name()))
		}
	}
	return nil
}

// StringToIntSlice converts string to []int
func StringToIntSlice(s, sep string) []int {
	if s == "" {
		return nil
	}

	fields := strings.Split(s, sep)
	slice := make([]int, len(fields))
	for i, f := range fields {
		slice[i] = ToInt(f)
	}
	return slice
}

// StringToInt32Slice converts string to []int32
func StringToInt32Slice(s, sep string) []int32 {
	if s == "" {
		return nil
	}

	fields := strings.Split(s, sep)
	slice := make([]int32, len(fields))
	for i, f := range fields {
		slice[i] = ToInt32(f)
	}
	return slice
}

// StringToInt64Slice converts string to []int64
func StringToInt64Slice(s, sep string) []int64 {
	if s == "" {
		return nil
	}

	fields := strings.Split(s, sep)
	slice := make([]int64, len(fields))
	for i, f := range fields {
		slice[i] = ToInt64(f)
	}
	return slice
}

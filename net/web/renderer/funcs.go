package renderer

import "time"

func Slice(values ...interface{}) interface{} {
	return values
}

func Map(pairs ...interface{}) interface{} {
	length := len(pairs)
	if length == 0 {
		return nil
	} else if length%2 != 0 {
		panic("Map function expect key-value pairs as arguments")
	}

	m := make(map[interface{}]interface{})
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

func Limit(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

func Range(i, min, max int) int {
	if i < min {
		return min
	} else if i > max {
		return max
	}
	return i
}

//func Range1(i, min, max interface{}) interface{} {
//	switch i.(type) {
//	case int, int8, int16, int32, int64:
//		if i < min {
//			return min
//		} else if i > max {
//			return max
//		}
//	}
//	return i
//}

func Time(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05")
}

func Date(t time.Time) string {
	return t.Local().Format("2006-01-02")
}

func Period(t time.Time) string {
	now := time.Now()
	if now.After(t) {
		d := now.Sub(t)
		return d.String() + " ago"
	} else if now.Before(t) {
		d := t.Sub(now)
		return d.String() + " later"
	}
	return "now"
}

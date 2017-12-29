package assert

import (
	"reflect"
	"testing"
	"time"

	"github.com/cuigh/auxo/util/cast"
)

// Same asserts that two objects are same.
func Same(tb testing.TB, expected, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if expected != actual {
		msg := format(msgAndArgs, "Not same: %p (expected) != %p (actual)", expected, actual)
		tb.Fatal(msg)
	}
}

// Equal asserts that two objects are equal.
func Equal(tb testing.TB, expected, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if !reflect.DeepEqual(expected, actual) {
		msg := format(msgAndArgs, "Not equal: %#v (expected) != %#v (actual)", expected, actual)
		tb.Fatal(msg)
	}
	//if expected == nil || actual == nil {
	//	return expected == actual
	//}
	//
	//return reflect.DeepEqual(expected, actual)
}

// NotEqual asserts that two objects are not equal.
func NotEqual(tb testing.TB, expected, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if reflect.DeepEqual(expected, actual) {
		msg := format(msgAndArgs, "Equal: %#v (expected) == %#v (actual)", expected, actual)
		tb.Fatal(msg)
	}
}

// Empty asserts that the specified object is empty.  I.e. nil, "", false, 0 or either
// a slice or a channel with len == 0.
func Empty(tb testing.TB, v interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if !isEmpty(v) {
		msg := format(msgAndArgs, "Should be empty, but was %v", v)
		tb.Fatal(msg)
	}
}

// NotEmpty asserts that the specified object is not empty.
func NotEmpty(tb testing.TB, v interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if isEmpty(v) {
		msg := format(msgAndArgs, "Should not be empty")
		tb.Fatal(msg)
	}
}

// True asserts that the specified value is true.
func True(tb testing.TB, v bool, msgAndArgs ...interface{}) {
	tb.Helper()

	if v != true {
		msg := format(msgAndArgs, "Should be true")
		tb.Fatal(msg)
	}
}

// False asserts that the specified value is false.
func False(tb testing.TB, v bool, msgAndArgs ...interface{}) {
	tb.Helper()

	if v != false {
		msg := format(msgAndArgs, "Should be false")
		tb.Fatal(msg)
	}
}

// Nil asserts that the specified object is nil.
func Nil(tb testing.TB, v interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if !isNil(v) {
		msg := format(msgAndArgs, "Expected nil, but got: %#v", v)
		tb.Fatal(msg)
	}
}

// NotNil asserts that the specified object is not nil.
func NotNil(tb testing.TB, v interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	if isNil(v) {
		msg := format(msgAndArgs, "Expected value not to be nil")
		tb.Fatal(msg)
	}
}

// Error asserts that the specified error is not nil.
func Error(tb testing.TB, err error, msgAndArgs ...interface{}) {
	tb.Helper()

	if isNil(err) {
		msg := format(msgAndArgs, "Should has error")
		tb.Fatal(msg)
	}
}

// NoError asserts that the specified error is nil.
func NoError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	tb.Helper()

	if !isNil(err) {
		msg := format(msgAndArgs, "Should no error, but got: %s", err)
		tb.Fatal(msg)
	}
}

// Contains asserts that the specified string, list(array, slice...) or map contains the
// specified substring or element.
//
//    assert.Contains(t, "Hello World", "World", "But 'Hello World' does contain 'World'")
//    assert.Contains(t, ["Hello", "World"], "World", "But ["Hello", "World"] does contain 'World'")
//    assert.Contains(t, {"Hello": "World"}, "Hello", "But {'Hello': 'World'} does contain 'Hello'")
func Contains(tb testing.TB, container interface{}, item interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	ok, found := contains(container, item)
	if !ok {
		msg := format(msgAndArgs, "'%s' is not string/array/slice/map type", container)
		tb.Fatal(msg)
	}
	if !found {
		msg := format(msgAndArgs, "'%s' does not contain '%s'", container, item)
		tb.Fatal(msg)
	}
}

// NotContains asserts that the specified string, list(array, slice...) or map does NOT contain the
// specified substring or element.
//
//    assert.NotContains(t, "Hello World", "Earth", "But 'Hello World' does NOT contain 'Earth'")
//    assert.NotContains(t, ["Hello", "World"], "Earth", "But ['Hello', 'World'] does NOT contain 'Earth'")
//    assert.NotContains(t, {"Hello": "World"}, "Earth", "But {'Hello': 'World'} does NOT contain 'Earth'")
func NotContains(tb testing.TB, container interface{}, item interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	ok, found := contains(container, item)
	if !ok {
		msg := format(msgAndArgs, "'%s' is not string/array/slice/map type", container)
		tb.Fatal(msg)
	}
	if found {
		msg := format(msgAndArgs, "'%s' should not contain '%s'", container, item)
		tb.Fatal(msg)
	}
}

// Panic asserts that the code inside the specified PanicTestFunc panics.
//
//   assert.Panic(t, func(){
//     Fatal()
//   }, "Calling Fatal() should panic")
func Panic(tb testing.TB, f func(), msgAndArgs ...interface{}) {
	tb.Helper()

	defer func() {
		if err := recover(); err != nil {
		}
	}()

	f()
	msg := format(msgAndArgs, "func %#v should panic", f)
	tb.Fatal(msg)
}

// NotPanic asserts that the code inside the specified PanicTestFunc does NOT panic.
//
//   assert.NotPanic(t, func(){
//     Safe()
//   }, "Calling Safe() should not panic")
func NotPanic(tb testing.TB, f func(), msgAndArgs ...interface{}) {
	tb.Helper()

	defer func() {
		if err := recover(); err != nil {
			msg := format(msgAndArgs, "func %#v should NOT panic", f)
			tb.Fatal(msg)
		}
	}()

	f()
}

// Implement asserts that the object implements the interface type u.
//
//   Implement(t, &bytes.Buffer{}, (*io.Writer)(nil))
func Implement(tb testing.TB, v, iface interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	it := reflect.TypeOf(iface).Elem()
	if !reflect.TypeOf(v).Implements(it) {
		msg := format(msgAndArgs, "%T must implement %v", v, it)
		tb.Fatal(msg)
	}
}

// Range asserts that the object is in range of two values.
func Range(tb testing.TB, actual, start, end interface{}, msgAndArgs ...interface{}) {
	tb.Helper()

	var ok bool
	switch v := actual.(type) {
	case time.Time:
		ok = !v.Before(start.(time.Time)) && !v.After(end.(time.Time))
	case int, int8, int16, int32, int64:
		i := cast.ToInt64(v)
		ok = i >= cast.ToInt64(start) && i <= cast.ToInt64(end)
	case uint, uint8, uint16, uint32, uint64:
		i := cast.ToUint64(v)
		ok = i >= cast.ToUint64(start) && i <= cast.ToUint64(end)
	case float32, float64:
		f := cast.ToFloat64(v)
		ok = f >= cast.ToFloat64(start) && f <= cast.ToFloat64(end)
	case string:
		ok = v >= start.(string) && v <= end.(string)
	}
	if !ok {
		msg := format(msgAndArgs, "Not in range: [%#v, %#v] (expected) != %#v (actual)", start, end, actual)
		tb.Fatal(msg)
	}
}

package assert

import (
	"fmt"
	"reflect"
	"strings"
)

// isNil checks if a specified object is nil or not, without Failing.
func isNil(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && v.IsNil() {
		return true
	}

	return false
}

// isEmpty checks if a value should be considered empty.
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
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
		return v.IsNil()
	}
	return reflect.DeepEqual(value, reflect.Zero(v.Type()).Interface())
}

// contains try loop over the list check if the list includes the element.
// return (false, false) if impossible.
// return (true, false) if element was not found.
// return (true, true) if element was found.
func contains(container interface{}, item interface{}) (ok, found bool) {
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	listValue := reflect.ValueOf(container)
	elementValue := reflect.ValueOf(item)

	if reflect.TypeOf(container).Kind() == reflect.String {
		return true, strings.Contains(listValue.String(), elementValue.String())
	}
	if reflect.TypeOf(container).Kind() == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if reflect.DeepEqual(mapKeys[i].Interface(), item) {
				return true, true
			}
		}
		return true, false
	}
	for i := 0; i < listValue.Len(); i++ {
		if reflect.DeepEqual(listValue.Index(i).Interface(), item) {
			return true, true
		}
	}
	return true, false
}

func format(msgAndArgs []interface{}, format string, args ...interface{}) string {
	switch len(msgAndArgs) {
	case 0:
		if len(args) == 0 {
			return format
		} else {
			return fmt.Sprintf(format, args...)
		}
	case 1:
		return msgAndArgs[0].(string)
	default:
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

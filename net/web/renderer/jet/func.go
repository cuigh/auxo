package jet

import (
	"reflect"

	"github.com/CloudyKit/jet"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/util/cast"
)

func equal(args jet.Arguments) reflect.Value {
	args.RequireNumOfArguments("eq", 2, 2)

	v1 := reflects.Indirect(args.Get(0))
	v2 := reflects.Indirect(args.Get(1))
	v, err := cast.TryToValue(v2.Interface(), v1.Type())
	if err != nil {
		return reflect.ValueOf(false)
	}
	return reflect.ValueOf(v1.Interface() == v.Interface())
}

// choose returns first non-empty argument or the last argument if not found.
func choose(args jet.Arguments) reflect.Value {
	args.RequireNumOfArguments("choose", 2, -1)

	var v reflect.Value
	for i, n := 0, args.NumOfArguments(); i < n; i++ {
		v = args.Get(i)
		if !reflects.IsEmpty(v) {
			return reflects.Indirect(v)
		}
	}
	return reflects.Indirect(v)
}

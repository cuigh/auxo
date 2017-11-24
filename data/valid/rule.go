package valid

import (
	"fmt"
	"math"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
)

var (
	rules = map[string]Rule{
		//"width":    nil, // string width, for CJK character, width is 2
		//"count":    nil, // rune count
		"required": requiredRule,
		"length":   lengthRule,
		"range":    rangeRule,
		"email":    matchRule("email", regexEmail),
		"url":      matchRule("url", regexURL),
		"alpha":    matchRule("alpha", regexAlpha),
		"regex":    regexRule,
		"ip":       ipRule,
		"ipv4":     ipv4Rule,
		"ipv6":     ipv6Rule,
	}
	messages = map[string]string{
		"$kind":    "rule `${rule}` can't apply to field `${name}`(${type})",
		"required": "field `${name}` is required",
		"length1":  "length must be ${argv}(not ${value})",
		"length2":  "length must in range ${arg}(not ${value})",
		"range1":   "value must be ${argv}(not ${value})",
		"range2":   "value must in range ${arg}(not %v)",
		"email":    "`${value}` is not a valid email address",
		"url":      "`${value}` is not a valid URL address",
		"alpha":    "`${value}` contains non-alpha letters",
		"regex":    "`${value}` doesn't match regex `${argv}`",
		"ip":       "`${value}` is not a valid IP address",
		"ipv4":     "`${value}` is not a valid IPV4 address",
		"ipv6":     "`${value}` is not a valid IPV6 address",
	}
)

type ArgumentFlag byte

const (
	LeftOpen    = '('
	LeftSquare  = '['
	RightOpen   = ')'
	RightSquare = ']'
)

type Context struct {
	Value    reflect.Value
	Info     *reflect.StructField
	messages map[string]string
}

type Argument struct {
	Value       string
	Left, Right ArgumentFlag
}

func (arg *Argument) String() string {
	return string(arg.Left) + arg.Value + string(arg.Right)
}

type Rule func(ctx *Context, arg *Argument) error

func requiredRule(ctx *Context, arg *Argument) error {
	v := reflects.Indirect(ctx.Value)
	ok := v.IsValid()
	if ok {
		switch v.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			ok = v.Len() > 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ok = v.Int() != 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ok = v.Uint() != 0
		case reflect.Float32, reflect.Float64:
			ok = v.Float() != 0
		default:
			ok = true
		}
	}
	return assert(ok, "required", ctx, arg, v.Interface())
}

func matchRule(name string, r *regexp.Regexp) Rule {
	return func(ctx *Context, arg *Argument) error {
		v := reflects.Indirect(ctx.Value)
		if err := assertKind(name, ctx, v, reflect.String); err != nil {
			return err
		}

		s := v.String()
		return assert(r.MatchString(s), name, ctx, arg, s)
	}
}

func ipRule(ctx *Context, arg *Argument) error {
	v := reflects.Indirect(ctx.Value)
	if err := assertKind("ip", ctx, v, reflect.String); err != nil {
		return err
	}

	s := v.String()
	ip := net.ParseIP(s)
	return assert(ip != nil, "ip", ctx, arg, s)
}

func ipv4Rule(ctx *Context, arg *Argument) error {
	v := reflects.Indirect(ctx.Value)
	if err := assertKind("ipv4", ctx, v, reflect.String); err != nil {
		return err
	}

	s := v.String()
	ip := net.ParseIP(s)
	return assert(ip != nil && strings.Contains(s, "."), "ipv4", ctx, arg, s)
}

func ipv6Rule(ctx *Context, arg *Argument) error {
	v := reflects.Indirect(ctx.Value)
	if err := assertKind("ipv6", ctx, v, reflect.String); err != nil {
		return err
	}

	s := v.String()
	ip := net.ParseIP(s)
	return assert(ip != nil && strings.Contains(s, ":"), "ipv6", ctx, arg, s)
}

func regexRule(ctx *Context, arg *Argument) error {
	if arg.Value == "" {
		return errors.Format("rule `regex` expected a pattern argument like: (\\d+) etc")
	}

	v := reflects.Indirect(ctx.Value)
	if err := assertKind("regex", ctx, v, reflect.String); err != nil {
		return err
	}

	s := v.String()
	ok, _ := regexp.MatchString(arg.Value, s)
	return assert(ok, "regex", ctx, arg, s)
}

// lengthRule can apply to string/slice/map/chan type.
func lengthRule(ctx *Context, arg *Argument) error {
	if arg.Value == "" || arg.Value == "~" {
		return errors.Format("rule `length` expected 1-2 arguments like: (5), [3~8) etc")
	}

	pair := strings.Split(arg.Value, "~")
	if len(pair) > 2 {
		return errors.Format("invalid argument for rule `length`: %v", arg.Value)
	}

	v := reflects.Indirect(ctx.Value)
	if err := assertKind("length", ctx, v, reflect.String, reflect.Slice, reflect.Map, reflect.Chan); err != nil {
		return err
	}

	l := v.Len()
	if len(pair) == 1 {
		length, err := strconv.Atoi(arg.Value)
		if err != nil {
			return err
		}
		return assert(l == length, "length1", ctx, arg, l)
	} else if len(pair) == 2 {
		start, err := toInt(pair[0], -1)
		if err != nil {
			return err
		}
		end, err := toInt(pair[1], math.MaxInt32)
		if err != nil {
			return err
		}

		if arg.Left == LeftOpen {
			err = assert(l > start, "length2", ctx, arg, l)
		} else if arg.Left == LeftSquare {
			err = assert(l >= start, "length2", ctx, arg, l)
		}
		if err != nil {
			return err
		}

		if arg.Right == RightOpen {
			err = assert(l < end, "length2", ctx, arg, l)
		} else if arg.Right == RightSquare {
			err = assert(l <= end, "length2", ctx, arg, l)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// rangeRule can apply to number types.
func rangeRule(ctx *Context, arg *Argument) error {
	if arg.Value == "" || arg.Value == "," {
		return errors.Format("rule `range` expected 1-2 arguments like: (5), [3~8) etc")
	}

	pair := strings.Split(arg.Value, "~")
	if len(pair) > 2 {
		return errors.Format("invalid argument for rule `range`: %v", arg.Value)
	}

	v := reflects.Indirect(ctx.Value)
	if err := assertKind("range", ctx, v, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64); err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := v.Int()
		if len(pair) == 1 {
			expected, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return err
			}
			return assert(value == expected, "range1", ctx, arg, value)
		} else {
			start, err := toInt64(pair[0], -1)
			if err != nil {
				return err
			}
			end, err := toInt64(pair[1], int64(math.MaxInt32))
			if err != nil {
				return err
			}

			if arg.Left == LeftOpen {
				err = assert(value > start, "range2", ctx, arg, value)
			} else if arg.Left == LeftSquare {
				err = assert(value >= start, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}

			if arg.Right == RightOpen {
				err = assert(value < end, "range2", ctx, arg, value)
			} else if arg.Right == RightSquare {
				err = assert(value <= end, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := v.Uint()
		if len(pair) == 1 {
			expected, err := strconv.ParseUint(arg.Value, 10, 64)
			if err != nil {
				return err
			}
			return assert(value == expected, "range1", ctx, arg, value)
		} else {
			start, err := toUint64(pair[0], 0)
			if err != nil {
				return err
			}
			end, err := toUint64(pair[1], uint64(math.MaxUint32))
			if err != nil {
				return err
			}

			if arg.Left == LeftOpen {
				err = assert(value > start, "range2", ctx, arg, value)
			} else if arg.Left == LeftSquare {
				err = assert(value >= start, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}

			if arg.Right == RightOpen {
				err = assert(value < end, "range2", ctx, arg, value)
			} else if arg.Right == RightSquare {
				err = assert(value <= end, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}
		}
	case reflect.Float32, reflect.Float64:
		value := v.Float()
		if len(pair) == 1 {
			expected, err := strconv.ParseFloat(arg.Value, 64)
			if err != nil {
				return err
			}
			return assert(value == expected, "range1", ctx, arg, value)
		} else {
			start, err := toFloat64(pair[0], -1)
			if err != nil {
				return err
			}
			end, err := toFloat64(pair[1], float64(math.MaxUint32))
			if err != nil {
				return err
			}

			if arg.Left == LeftOpen {
				err = assert(value > start, "range2", ctx, arg, value)
			} else if arg.Left == LeftSquare {
				err = assert(value >= start, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}

			if arg.Right == RightOpen {
				err = assert(value < end, "range2", ctx, arg, value)
			} else if arg.Right == RightSquare {
				err = assert(value <= end, "range2", ctx, arg, value)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func toInt(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}

func toInt64(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func toUint64(s string, def uint64) (uint64, error) {
	if s == "" {
		return def, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func toFloat64(s string, def float64) (float64, error) {
	if s == "" {
		return def, nil
	}
	return strconv.ParseFloat(s, 64)
}

func assertKind(rule string, ctx *Context, value reflect.Value, kinds ...reflect.Kind) error {
	for _, kind := range kinds {
		if kind == value.Kind() {
			return nil
		}
	}

	msg := ctx.messages["$kind"]
	return errors.New(os.Expand(msg, func(name string) string {
		switch name {
		case "rule":
			return rule
		case "name":
			return ctx.Info.Name
		case "type":
			return value.Type().String()
		default:
			return ""
		}
	}))
}

func assert(ok bool, name string, ctx *Context, arg *Argument, value interface{}) error {
	if !ok {
		return failed(name, ctx, arg, value)
	}
	return nil
}

func failed(name string, ctx *Context, arg *Argument, value interface{}) error {
	msg := ctx.messages[name]
	if msg == "" {
		msg = messages[name]
	}
	if msg == "" {
		panic("can not find message for: " + name)
	}

	return errors.New(os.Expand(msg, func(name string) string {
		switch name {
		case "name":
			return ctx.Info.Name
		case "value":
			return fmt.Sprint(value)
		case "arg":
			return arg.String()
		case "argv":
			return arg.Value
		default:
			return "[" + name + "]"
		}
	}))
}

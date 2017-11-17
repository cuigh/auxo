package valid

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
)

var (
	// Tag is the default tag key for validation.
	Tag       = "valid"
	validator = new(Validator)
)

// Register adds a rule to default validator.
func Register(name string, rule Rule) {
	rules[name] = rule
}

// Prompt sets the default validation message for rules.
func Prompt(name, msg string) {
	messages[name] = msg
}

// Validate checks value of struct is available.
func Validate(i interface{}) error {
	return validator.Validate(i)
}

type Error struct {
	Field string
	Rule  string
	cause error
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Error() string {
	return fmt.Sprintf("failed to validate field `%s` with rule `%s`: %v", e.Field, e.Rule, e.cause)
}

// Validator is a struct data validator.
type Validator struct {
	Tag      string
	rules    map[string]Rule
	messages map[string]string
}

// Register adds a rule to validator.
func (v *Validator) Register(name string, rule Rule) {
	if v.rules == nil {
		v.rules = make(map[string]Rule)
	}
	v.rules[name] = rule
}

// Prompt sets the validation message for rules.
func (v *Validator) Prompt(name, msg string) {
	if v.messages == nil {
		v.messages = make(map[string]string)
	}
	v.messages[name] = msg
}

// Validate checks value of struct is available.
func (v *Validator) Validate(i interface{}) error {
	value := reflects.Indirect(reflect.ValueOf(i))
	if value.Kind() != reflect.Struct {
		return errors.New("valid: target value must be a struct")
	}

	key := v.Tag
	if key == "" {
		key = Tag
	}
	ctx := &Context{messages: messages}
	sv := reflects.StructOf(value)
	return sv.VisitFields(func(fv reflect.Value, fi *reflect.StructField) error {
		rules, err := v.parseRules(fi.Tag.Get(key))
		if err != nil {
			return err
		}

		if len(rules) == 0 {
			fv = reflects.Indirect(fv)
			if fv.Kind() == reflect.Struct && fv.IsValid() {
				return v.Validate(fv.Interface())
			}
			return nil
		}

		ctx.Value, ctx.Info = fv, fi
		for name, info := range rules {
			if r := v.getRule(name); r != nil {
				if err = r(ctx, &info); err != nil {
					return &Error{Field: fi.Name, Rule: name, cause: err}
				}
			} else {
				return errors.New("unknown rule: " + name)
			}
		}
		return nil
	})
}

func (v *Validator) getRule(name string) Rule {
	if v.rules != nil {
		if r, ok := v.rules[name]; ok {
			return r
		}
	}
	return rules[name]
}

func (v *Validator) parseRules(tag string) (rules map[string]Argument, err error) {
	// tag: length[1,5),regex(\d+),ip,width(5)
	const (
		stateName = 0
		//stateLeftFlag  = 1
		stateRightFlag = 2
		stateArg       = 3
	)

	var (
		state       = stateName
		last        byte
		name        string
		left, right ArgumentFlag
		buf         bytes.Buffer
	)

	rules = make(map[string]Argument)
	for i, l := 0, len(tag); i < l; i++ {
		b := tag[i]
		switch state {
		case stateName:
			if b == LeftOpen || b == LeftSquare {
				state = stateArg
				left = ArgumentFlag(b)
				name = buf.String()
				buf.Reset()
			} else if b == ',' {
				rules[buf.String()] = Argument{}
				buf.Reset()
			} else {
				buf.WriteByte(b)
			}
		case stateArg:
			if b == RightOpen || b == RightSquare {
				state = stateRightFlag
				right = ArgumentFlag(b)
			} else {
				buf.WriteByte(b)
			}
		case stateRightFlag:
			if b == ',' {
				state = stateName
				rules[name] = Argument{Value: buf.String(), Left: left, Right: right}
				buf.Reset()
			} else if (b == RightOpen || b == RightSquare) && last == b {
				state = stateArg
				buf.WriteByte(b)
			} else {
				return nil, fmt.Errorf("invalid rule: %s(pos: %d)", tag, i)
			}
		}
		last = tag[i]
	}

	if state == stateName {
		if buf.Len() > 0 {
			rules[buf.String()] = Argument{}
		}
	} else if state == stateRightFlag {
		rules[name] = Argument{Value: buf.String(), Left: left, Right: right}
	} else {
		return nil, fmt.Errorf("invalid rule: %s(pos: %d)", tag, len(tag)-1)
	}
	return
}

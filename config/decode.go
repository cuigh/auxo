package config

import (
	"reflect"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/util/cast"
)

// Unmarshal exports options to struct.
func (m *Manager) Unmarshal(v interface{}) error {
	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		panic("v must be a pointer of struct")
	}

	vv := reflect.ValueOf(v).Elem()
	if vv.Kind() != reflect.Struct {
		panic("v must be a pointer of struct")
	}

	return unmarshal(vv, m.Get)
}

// Unmarshal exports specific option to struct.
func (m *Manager) UnmarshalOption(name string, v interface{}) error {
	value := m.Get(name)
	if value == nil {
		return errors.Format("option [%v] is not found", name)
	}

	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer")
	}

	vv := reflect.ValueOf(v).Elem()
	return unmarshalValue(vv, value)
}

// unmarshal to struct value
func unmarshal(v reflect.Value, valuer func(name string) interface{}) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i, num := 0, t.NumField(); i < num; i++ {
		var (
			f    = v.Field(i)
			sf   = t.Field(i)
			name = sf.Tag.Get("option") // todo: support default value
		)

		if sf.PkgPath != "" {
			continue
		}

		if name == "" {
			name = texts.Rename(sf.Name, texts.Lower)
		}
		opt := valuer(name)
		if opt == nil {
			// todo: read default value with option tag?
			continue
		}

		if err := unmarshalValue(f, opt); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalValue(v reflect.Value, value interface{}) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.CanAddr() && v.Addr().Type().Implements(unmarshalerType) {
		return v.Addr().Interface().(Unmarshaler).Unmarshal(value)
	}

	if isSimpleType(v.Type()) {
		return unmarshalSimpleValue(v, value)
	}

	switch v.Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(value))
		return nil
	case reflect.Array:
		return errors.NotImplemented
	case reflect.Slice:
		return unmarshalSliceValue(v, value)
	case reflect.Struct:
		if m, ok := tryConvertMap(value); ok {
			return unmarshal(v, m.Find)
		}
		return decodeError(value, v.Type())
	case reflect.Map:
		return unmarshalMapValue(v, value)
	default:
		return decodeError(value, v.Type())
	}
}

func isSimpleType(t reflect.Type) bool {
	if t == reflects.TypeTime || t == reflects.TypeDuration {
		return true
	}

	switch t.Kind() {
	case reflect.Bool, reflect.String, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func unmarshalSimpleValue(f reflect.Value, opt interface{}) error {
	if v, err := cast.TryToValue(opt, f.Type()); err != nil {
		return decodeError(opt, f.Type(), err)
	} else {
		f.Set(v)
		return nil
	}
}

func unmarshalSliceValue(f reflect.Value, opt interface{}) error {
	et := f.Type().Elem()
	if isSimpleType(et) {
		return unmarshalSimpleValue(f, opt)
	}

	v := reflect.ValueOf(opt)
	if v.Kind() != reflect.Slice {
		return decodeError(opt, f.Type())
	}

	l := v.Len()
	if l == 0 {
		return nil
	}

	sv := reflects.SliceOf(f)
	for i := 0; i < l; i++ {
		x := reflect.New(et).Elem()
		opt = v.Index(i).Interface()
		err := unmarshalValue(x, opt)
		if err != nil {
			return err
		}
		sv.AddValue(x)
	}
	return nil
}

func unmarshalMapValue(f reflect.Value, opt interface{}) error {
	m, ok := tryConvertMap(opt)
	if !ok {
		return decodeError(opt, f.Type())
	}

	if len(m) == 0 {
		return nil
	}

	mv := reflects.MapOf(f)
	for key, value := range m {
		// Here we assume key type is simple.
		kv, err := cast.TryToValue(key, f.Type().Key())
		if err != nil {
			return err
		}

		vv := reflect.New(f.Type().Elem()).Elem()
		if err = unmarshalValue(vv, value); err != nil {
			return err
		}
		mv.SetValue(kv, vv)
	}
	return nil
}

func tryConvertMap(i interface{}) (data.Map, bool) {
	switch v := i.(type) {
	case data.Map:
		return v, true
	case map[string]interface{}:
		return data.Map(v), true
	}
	return nil, false
}

func decodeError(opt interface{}, target reflect.Type, err ...error) error {
	if len(err) > 0 {
		return errors.Format("can't decode option '%#v' to type %v: %v", opt, target, err[0])
	}
	return errors.Format("can't decode option '%#v' to type %v", opt, target)
}

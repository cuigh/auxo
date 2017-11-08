package data

import (
	"strings"
)

type Map map[string]interface{}

func (m Map) Get(key string) interface{} {
	return m[key]
}

func (m Map) TryGet(key string) (v interface{}, ok bool) {
	v, ok = m[key]
	return
}

func (m Map) Set(key string, value interface{}) Map {
	m[key] = value
	return m
}

func (m Map) Remove(key string) {
	delete(m, key)
}

func (m Map) Keys() []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func (m Map) Contains(key string) bool {
	_, ok := m[key]
	return ok
}

// Merge merges all key-value pairs in m from src.
func (m Map) Merge(src Map) {
	for k, sv := range src {
		if tv, ok := m[k]; ok {
			m1, ok1 := m.tryConvert(tv)
			m2, ok2 := m.tryConvert(sv)
			if ok1 && ok2 {
				m1.Merge(m2)
			}
		} else {
			if tm, ok := m.tryConvert(sv); ok {
				m[k] = tm
			} else {
				m[k] = sv
			}
		}
	}
}

// Cover merges and replaces all key-value pairs in m from src
func (m Map) Cover(src Map) {
	for k, sv := range src {
		if tv, ok := m[k]; ok {
			m1, ok1 := m.tryConvert(tv)
			m2, ok2 := m.tryConvert(sv)
			if ok1 && ok2 {
				m1.Cover(m2)
				continue
			}
		}

		if tm, ok := m.tryConvert(sv); ok {
			m[k] = tm
		} else {
			m[k] = sv
		}
	}
}

func (m Map) Find(key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}

	for i := strings.LastIndex(key, "."); i != -1; i = strings.LastIndex(key[:i], ".") {
		if v, ok := m[key[:i]]; ok {
			if tmp, ok := m.tryConvert(v); ok {
				return tmp.Find(key[i+1:])
			}
		}
	}
	return nil
}

//func (m Map) Unmarshal(i interface{}, namer func(sf *reflect.StructField) string, valuer func(f reflect.Value, i interface{}) (bool, error)) error {
//	v := reflect.ValueOf(i)
//	if v.Kind() == reflect.Ptr {
//		v = v.Elem()
//	}
//
//	if v.Kind() == reflect.Struct {
//		return errors.New("i must be a struct pointer")
//	}
//
//	t := v.Type()
//	for i, n := 0, v.NumField(); i < n; i++ {
//		f := v.Field(i)
//		sf := t.Field(i)
//		name := namer(&sf)
//		value := m.Find(name)
//		if value == nil {
//			continue
//		}
//
//		if f.Kind() == reflect.Ptr {
//			if f.IsNil() {
//				f.Set(reflect.New(f.Type().Elem()))
//			}
//			f = f.Elem()
//		}
//
//		if valuer != nil {
//			ok, err := valuer(f, value)
//			if ok {
//				continue
//			} else if err != nil {
//				return err
//			}
//		}
//
//	}
//	return nil
//}

func (m Map) tryConvert(i interface{}) (Map, bool) {
	switch v := i.(type) {
	case Map:
		return v, true
	case map[string]interface{}:
		return Map(v), true
		//case []interface{}:
		//	for i, value := range v {
		//		if tm, ok := m.tryConvert(value); ok {
		//			v[i] = tm
		//		}
		//	}
	}
	return nil, false
}

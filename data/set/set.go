package set

import (
	"reflect"
)

type Set map[interface{}]struct{}

func (s Set) Add(items ...interface{}) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set) AddSlice(slice interface{}, fn func(i int) interface{}) {
	if slice == nil {
		return
	}

	length := reflect.ValueOf(slice).Len()
	for i := 0; i < length; i++ {
		s[fn(i)] = struct{}{}
	}
}

func (s Set) Remove(items ...interface{}) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set) RemoveSlice(slice interface{}, fn func(i int) string) {
	if slice == nil || len(s) == 0 {
		return
	}

	length := reflect.ValueOf(slice).Len()
	for i := 0; i < length; i++ {
		delete(s, fn(i))
	}
}

func (s Set) Contains(item interface{}) bool {
	_, ok := s[item]
	return ok
}

func (s Set) Union(set Set) {
	for k := range set {
		s[k] = struct{}{}
	}
}

func (s Set) Intersect(set Set) {
	for k := range set {
		if !set.Contains(k) {
			delete(s, k)
		}
	}
}

func (s Set) Len() int {
	return len(s)
}

//func (s Set) Slice() interface{} {
//	v := reflect.MakeSlice(reflect.Type, s.Len(), s.Len())
//	return v.Interface()
//}

func NewSet(items ...interface{}) Set {
	s := Set{}
	s.Add(items...)
	return s
}

func Union(sets ...Set) Set {
	s := Set{}
	for _, set := range sets {
		s.Union(set)
	}
	return s
}

func Intersect(sets ...Set) Set {
	s := Set{}
	for _, set := range sets {
		s.Intersect(set)
	}
	return s
}

type StringSet map[string]struct{}

func NewStringSet(items ...string) StringSet {
	s := StringSet{}
	s.Add(items...)
	return s
}

func (s StringSet) Add(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s StringSet) AddSlice(slice interface{}, fn func(i int) string) {
	if slice == nil {
		return
	}

	length := reflect.ValueOf(slice).Len()
	for i := 0; i < length; i++ {
		s[fn(i)] = struct{}{}
	}
}

func (s StringSet) Remove(items ...string) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s StringSet) RemoveSlice(slice interface{}, fn func(i int) string) {
	if slice == nil || len(s) == 0 {
		return
	}

	length := reflect.ValueOf(slice).Len()
	for i := 0; i < length; i++ {
		delete(s, fn(i))
	}
}

func (s StringSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s StringSet) Union(set StringSet) {
	for k := range set {
		s[k] = struct{}{}
	}
}

func (s StringSet) Intersect(set StringSet) {
	for k := range set {
		if !set.Contains(k) {
			delete(s, k)
		}
	}
}

func (s StringSet) Len() int {
	return len(s)
}

func (s StringSet) Slice() []string {
	slice := make([]string, 0, len(s))
	for k := range s {
		slice = append(slice, k)
	}
	return slice
}

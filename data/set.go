package data

import (
	"reflect"
)

// Empty is a empty struct instance.
var Empty = struct{}{}

type Set map[interface{}]struct{}

func (s Set) Add(items ...interface{}) {
	for _, item := range items {
		s[item] = Empty
	}
}

func (s Set) AddSlice(slice interface{}, fn func(i int) interface{}) {
	if slice == nil {
		return
	}

	length := reflect.ValueOf(slice).Len()
	for i := 0; i < length; i++ {
		s.Add(fn(i))
	}
}

func (s Set) Remove(item interface{}) {
	delete(s, item)
}

func (s Set) Contains(item interface{}) bool {
	_, ok := s[item]
	return ok
}

func (s Set) Union(set Set) {
	for k := range set {
		s[k] = Empty
	}
}

func (s Set) Len() int {
	return len(s)
}

//func (s Set) ToArray() interface{} {
//	v := reflect.MakeSlice(reflect.Type, s.Len(), s.Len())
//	return v.Interface()
//}

func NewSet(items ...interface{}) Set {
	s := Set{}
	s.Add(items...)
	return s
}

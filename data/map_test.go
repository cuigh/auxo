package data

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestMapMerge(t *testing.T) {
	m1 := Map{
		"a": 11,
		"b": Map{},
		"c": Map{
			"a": 11,
		},
	}
	m2 := Map{
		"a": "x",
		"b": 21,
		"c": Map{
			"a": "x",
			"b": 21,
		},
		"d": 24,
	}
	m1.Merge(m2)
	t.Log(m1)
}

func TestMapCover(t *testing.T) {
	m1 := Map{
		"a": 11,
		"b": Map{
			"a": 11,
			"b": 12,
		},
	}
	m2 := Map{
		"a": "x",
		"b": map[string]interface{}{
			"a": 21,
		},
		"c": 23,
	}
	m1.Cover(m2)

	cases := []struct {
		Key   string
		Value interface{}
	}{
		{"a", "x"},
		{"b.a", 21},
	}

	for _, c := range cases {
		assert.Equal(t, c.Value, m1.Find(c.Key))
	}
}

func TestMapFind(t *testing.T) {
	m := Map{
		"a": 11,
		"b": Map{
			"a": 11,
			"b": 12,
		},
		"b.b": "test",
	}

	cases := []struct {
		Key   string
		Value interface{}
	}{
		{"a", 11},
		{"b.a", 11},
		{"b.b", "test"},
	}

	for _, c := range cases {
		v := m.Find(c.Key)
		assert.Equal(t, c.Value, v)
	}
}

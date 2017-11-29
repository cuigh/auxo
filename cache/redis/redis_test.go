package redis

import (
	"reflect"
	"testing"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/test/assert"
)

const key = "key"

type User struct {
	Name string
}

func init() {
	config.AddFolder("..")
}

func TestProvider(t *testing.T) {
	var (
		value  = 10
		actual int
	)

	p, err := NewProvider(data.Map{"db": "cache"})
	assert.NoError(t, err)

	err = p.Set(key, value, time.Minute)
	assert.NoError(t, err)

	ok, err := p.Exist(key)
	assert.NoError(t, err)
	assert.True(t, ok)

	v, err := p.Get(key)
	assert.NoError(t, err)

	actual, err = v.Int()
	assert.NoError(t, err)
	assert.Equal(t, value, actual)

	err = p.Remove(key)
	assert.NoError(t, err)
}

func TestValue(t *testing.T) {
	testCases := []struct {
		Actual   interface{}
		IsNil    bool
		Expected interface{}
		New      func() interface{}
	}{
		{nil, true, nil, nil},
		{true, false, true, func() interface{} { return new(bool) }},
		{"test", false, "test", func() interface{} { return new(string) }},
		{[]byte("test"), false, []byte("test"), func() interface{} { return &[]byte{} }},
		{float32(1.5), false, float32(1.5), func() interface{} { return new(float32) }},
		{float64(1.5), false, float64(1.5), func() interface{} { return new(float64) }},
		{int(1), false, int(1), func() interface{} { return new(int) }},
		{int8(1), false, int8(1), func() interface{} { return new(int8) }},
		{int16(1), false, int16(1), func() interface{} { return new(int16) }},
		{int32(1), false, int32(1), func() interface{} { return new(int32) }},
		{int64(1), false, int64(1), func() interface{} { return new(int64) }},
		{User{"test"}, false, User{"test"}, func() interface{} { return new(User) }},
		{&User{"test"}, false, User{"test"}, func() interface{} { return new(User) }},
	}

	p, err := NewProvider(data.Map{"db": "cache"})
	assert.NoError(t, err)

	for _, tc := range testCases {
		err := p.Set(key, tc.Actual, time.Minute)
		assert.NoError(t, err)

		ok, err := p.Exist(key)
		assert.NoError(t, err)
		assert.True(t, ok)

		v, err := p.Get(key)
		assert.NoError(t, err)
		if tc.IsNil {
			assert.True(t, v.IsNil())
		} else {
			i := tc.New()
			err = v.Scan(i)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, reflect.ValueOf(i).Elem().Interface())
		}
	}
}

func BenchmarkProvider_Get(b *testing.B) {
	b.ReportAllocs()

	p, err := NewProvider(data.Map{"db": "cache"})
	assert.NoError(b, err)

	p.Set(key, 10, time.Minute)

	for i := 0; i < b.N; i++ {
		if _, err := p.Get(key); err != nil {
			b.Fatal(err)
		}
	}
}

package lazy_test

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/lazy"
)

func TestValue(t *testing.T) {
	expected := time.Now()
	value := lazy.Value{
		New: func() (interface{}, error) {
			return expected, nil
		},
	}

	v, err := value.Get()
	assert.NoError(t, err)
	assert.NotNil(t, v)

	dt := v.(time.Time)
	assert.Equal(t, expected, dt)
}

func BenchmarkValue(b *testing.B) {
	expected := time.Now()
	value := lazy.Value{
		New: func() (interface{}, error) {
			return expected, nil
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := value.Get()
		if err != nil {
			b.Fail()
		}
		if dt := v.(time.Time); !dt.Equal(expected) {
			b.Fail()
		}
	}
}

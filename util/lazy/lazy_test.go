package lazy_test

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/lazy"
)

func TestValue(t *testing.T) {
	expected := time.Now()
	value := lazy.Value[time.Time]{
		New: func() (time.Time, error) {
			return expected, nil
		},
	}

	v, err := value.Get()
	assert.NoError(t, err)
	assert.NotNil(t, v)

	assert.Equal(t, expected, v)
}

func BenchmarkValue(b *testing.B) {
	expected := time.Now()
	value := lazy.Value[time.Time]{
		New: func() (time.Time, error) {
			return expected, nil
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := value.Get()
		if err != nil {
			b.Fail()
		}
		if !v.Equal(expected) {
			b.Fail()
		}
	}
}

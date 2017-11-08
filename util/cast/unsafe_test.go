package cast_test

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/cast"
)

func TestStringToBytes(t *testing.T) {
	s := "auxo"
	b := cast.StringToBytes(s)
	assert.Equal(t, s, string(b))
}

func TestBytesToString(t *testing.T) {
	b := []byte("auxo")
	s := cast.BytesToString(b)
	assert.Equal(t, string(b), s)
}

func BenchmarkStringToBytes(b *testing.B) {
	s := "auxo"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cast.StringToBytes(s)
	}
}

func BenchmarkBytesToString(b *testing.B) {
	bytes := []byte("auxo")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cast.BytesToString(bytes)
	}
}

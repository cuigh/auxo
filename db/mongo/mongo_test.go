package mongo

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/test/assert"
)

func init() {
	config.AddFolder(".")
}

func TestFactory_Open(t *testing.T) {
	fn := func() {
		db := MustOpen("test")
		defer db.Close()

		_, err := db.C("user").Count()
		assert.NoError(t, err)
	}
	fn()
	fn()
	fn()
}

func BenchmarkOpen(b *testing.B) {
	fn := func() {
		db := MustOpen("test")
		defer db.Close()
	}
	fn()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fn()
	}
}

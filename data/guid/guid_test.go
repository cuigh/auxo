package guid

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestNew(t *testing.T) {
	id := New()
	assert.Equal(t, 20, len(id))
	t.Log(id)
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

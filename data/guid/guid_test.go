package guid

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestNew(t *testing.T) {
	id := New()
	assert.Equal(t, 20, len(id.String()))
	assert.Equal(t, 12, len(id.Slice()))
	t.Log(id)

	b, err := id.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, 20, len(b))
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

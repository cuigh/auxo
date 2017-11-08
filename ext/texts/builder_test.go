package texts

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestBuilder_Append(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.Append("1")
	b.Append("2", "3")
	assert.Equal(t, "123", b.String())
}

func TestBuilder_AppendFormat(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.AppendFormat("id=%v", 1)
	assert.Equal(t, "id=1", b.String())
}

func TestBuilder_AppendByte(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.AppendByte('1')
	assert.Equal(t, "1", b.String())
}

func TestBuilder_AppendBytes(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.AppendBytes('1', '2', '3')
	assert.Equal(t, "123", b.String())
}

func TestBuilder_Reset(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.Append("1").Reset()
	assert.Equal(t, "", b.String())
}

func TestBuilder_Truncate(t *testing.T) {
	b := GetBuilder()
	defer PutBuilder(b)

	b.Append("123").Truncate(2)
	assert.Equal(t, "12", b.String())
}

func BenchmarkBuilder(b *testing.B) {
	fn := func() {
		b := GetBuilder()
		b.Append("abc", "cde", "adfsdf", "sdfdsf")
		b.String()
		PutBuilder(b)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

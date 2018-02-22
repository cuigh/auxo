package texts

import (
	"bytes"
	"fmt"
	"sync"
	"unsafe"
)

var (
	builders = sync.Pool{
		New: func() interface{} {
			return &Builder{}
		},
	}
)

func GetBuilder() *Builder {
	return builders.Get().(*Builder)
}

func PutBuilder(b *Builder) {
	b.Reset()
	builders.Put(b)
}

// Builder is used to efficiently build a string using Write methods. It minimizes memory copying.
// The zero value is ready to use.
type Builder bytes.Buffer

func (b *Builder) Reset() *Builder {
	(*bytes.Buffer)(b).Reset()
	return b
}

func (b *Builder) Truncate(n int) *Builder {
	(*bytes.Buffer)(b).Truncate(n)
	return b
}

func (b *Builder) Write(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Write(p)
}

func (b *Builder) WriteByte(c byte) *Builder {
	(*bytes.Buffer)(b).WriteByte(c)
	return b
}

func (b *Builder) WriteString(s ...string) *Builder {
	buf := (*bytes.Buffer)(b)
	for _, e := range s {
		buf.WriteString(e)
	}
	return b
}

func (b *Builder) WriteStringer(s ...fmt.Stringer) *Builder {
	buf := (*bytes.Buffer)(b)
	for _, e := range s {
		buf.WriteString(e.String())
	}
	return b
}

func (b *Builder) WriteFormat(format string, args ...interface{}) *Builder {
	(*bytes.Buffer)(b).WriteString(fmt.Sprintf(format, args...))
	return b
}

func (b *Builder) Len() int {
	return (*bytes.Buffer)(b).Len()
}

func (b *Builder) Bytes() []byte {
	return (*bytes.Buffer)(b).Bytes()
}

func (b *Builder) String() string {
	buf := (*bytes.Buffer)(b).Bytes()
	return *(*string)(unsafe.Pointer(&buf))
}

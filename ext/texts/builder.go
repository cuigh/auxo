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

type Builder bytes.Buffer

func GetBuilder() *Builder {
	return builders.Get().(*Builder)
}

func PutBuilder(b *Builder) {
	b.Reset()
	builders.Put(b)
}

func (b *Builder) Write(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Write(p)
}

func (b *Builder) Reset() *Builder {
	(*bytes.Buffer)(b).Reset()
	return b
}

func (b *Builder) Truncate(n int) *Builder {
	(*bytes.Buffer)(b).Truncate(n)
	return b
}

func (b *Builder) Length() int {
	return (*bytes.Buffer)(b).Len()
}

func (b *Builder) Append(first string, others ...string) *Builder {
	(*bytes.Buffer)(b).WriteString(first)
	for _, s := range others {
		(*bytes.Buffer)(b).WriteString(s)
	}
	return b
}

func (b *Builder) AppendBytes(p ...byte) *Builder {
	(*bytes.Buffer)(b).Write(p)
	return b
}

func (b *Builder) AppendByte(c byte) *Builder {
	(*bytes.Buffer)(b).WriteByte(c)
	return b
}

func (b *Builder) AppendFormat(format string, args ...interface{}) *Builder {
	s := fmt.Sprintf(format, args...)
	(*bytes.Buffer)(b).WriteString(s)
	return b
}

func (b *Builder) Bytes() []byte {
	return (*bytes.Buffer)(b).Bytes()
}

func (b *Builder) String() string {
	buf := (*bytes.Buffer)(b).Bytes()
	return *(*string)(unsafe.Pointer(&buf))
	//return (*bytes.Buffer)(b).String()
}

package assert

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestEqual(t *testing.T) {
	Equal(t, nil, nil)
	Equal(t, 0, 0)
}

func TestEmpty(t *testing.T) {
	Empty(t, nil)
	Empty(t, 0)
	Empty(t, false)
	Empty(t, time.Time{})
}

func TestTrue(t *testing.T) {
	True(t, true)
}

func TestFalse(t *testing.T) {
	False(t, false)
}

func TestNil(t *testing.T) {
	Nil(t, nil)
}

func TestNotNil(t *testing.T) {
	NotNil(t, 0)
}

func TestContains(t *testing.T) {
	Contains(t, "Hello World", "World")
	Contains(t, []string{"Hello", "World"}, "World")
	Contains(t, map[string]string{"Hello": "World"}, "Hello")
}

func TestNotContains(t *testing.T) {
	NotContains(t, "Hello World", "Foo")
	NotContains(t, []string{"Hello", "World"}, "Foo")
	NotContains(t, map[string]string{"Hello": "World"}, "Foo")
}

func TestPanic(t *testing.T) {
	Panic(t, func() {
		panic("ooops..")
	})
}

func TestNotPanic(t *testing.T) {
	NotPanic(t, func() {
	})
}

func TestImplement(t *testing.T) {
	Implement(t, &bytes.Buffer{}, (*io.Writer)(nil))
}

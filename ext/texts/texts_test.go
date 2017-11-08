package texts

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

type TestCase struct {
	Input  string
	Expect string
}

func TestJoin(t *testing.T) {
	cases := []struct {
		Input  []string
		Expect string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a,b,c"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, Join(",", c.Input...))
	}
}

func TestJoinRepeat(t *testing.T) {
	s := "?"
	sep := ","
	cases := []struct {
		Count  int
		Expect string
	}{
		{0, ""},
		{1, "?"},
		{2, "?,?"},
		{3, "?,?,?"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, JoinRepeat(s, sep, c.Count))
	}
}

func TestConcat(t *testing.T) {
	cases := []struct {
		Input  []string
		Expect string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "abc"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, Concat(c.Input...))
	}
}

func TestPadLeft(t *testing.T) {
	width := 5
	var padding byte = '0'
	cases := []TestCase{
		{
			Input:  "1234567",
			Expect: "1234567",
		},
		{
			Input:  "12345",
			Expect: "12345",
		},
		{
			Input:  "123",
			Expect: "00123",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, PadLeft(c.Input, padding, width))
	}
}

func TestPadRight(t *testing.T) {
	width := 5
	var padding byte = '0'
	cases := []TestCase{
		{
			Input:  "1234567",
			Expect: "1234567",
		},
		{
			Input:  "12345",
			Expect: "12345",
		},
		{
			Input:  "123",
			Expect: "12300",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, PadRight(c.Input, padding, width))
	}
}

func TestPadCenter(t *testing.T) {
	width := 5
	var padding byte = '0'
	cases := []TestCase{
		{
			Input:  "1234567",
			Expect: "1234567",
		},
		{
			Input:  "12345",
			Expect: "12345",
		},
		{
			Input:  "123",
			Expect: "01230",
		},
		{
			Input:  "12",
			Expect: "01200",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, PadCenter(c.Input, padding, width))
	}
}

func TestCutLeft(t *testing.T) {
	width := 5
	cases := []TestCase{
		{
			Input:  "abcdefg",
			Expect: "cdefg",
		},
		{
			Input:  "abcde",
			Expect: "abcde",
		},
		{
			Input:  "abc",
			Expect: "abc",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, CutLeft(c.Input, width))
	}
}

func TestCutRight(t *testing.T) {
	width := 5
	cases := []TestCase{
		{
			Input:  "abcdefg",
			Expect: "abcde",
		},
		{
			Input:  "abcde",
			Expect: "abcde",
		},
		{
			Input:  "abc",
			Expect: "abc",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expect, CutRight(c.Input, width))
	}
}

func TestIsUpper(t *testing.T) {
	assert.True(t, IsUpper("ABC"))
	assert.False(t, IsUpper("ABc"))
}

func TestIsLower(t *testing.T) {
	assert.True(t, IsLower("abc"))
	assert.False(t, IsLower("abC"))
}

func TestRename(t *testing.T) {
	styles := []NameStyle{Pascal, Camel, Upper, Lower}
	cases := []struct {
		Input  string
		Expect []string
	}{
		{
			Input:  "name",
			Expect: []string{"Name", "name", "NAME", "name"},
		},
		{
			Input:  "NAME",
			Expect: []string{"Name", "name", "NAME", "name"},
		},
		{
			Input:  "user_name",
			Expect: []string{"UserName", "userName", "USER_NAME", "user_name"},
		},
		{
			Input:  "USER_NAME",
			Expect: []string{"UserName", "userName", "USER_NAME", "user_name"},
		},
		{
			Input:  "UserName",
			Expect: []string{"UserName", "userName", "USER_NAME", "user_name"},
		},
		{
			Input:  "userName",
			Expect: []string{"UserName", "userName", "USER_NAME", "user_name"},
		},
		//{
		//	Input:  "HTTPMethod",
		//	Expect: []string{"HTTPMethod", "httpMethod", "HTTP_METHOD", "http_method"},
		//},
		//{
		//	Input:  "myHTTPMethod_Get",
		//	Expect: []string{"MyHTTPMethodGet", "myHTTPMethodGet", "MY_HTTP_METHOD_GET", "my_http_method_get"},
		//},
	}
	for _, c := range cases {
		for i, style := range styles {
			r := Rename(c.Input, style)
			assert.Equal(t, c.Expect[i], r, "%v > %#v != %#v (%v)", c.Input, c.Expect[i], r, style)
		}
	}
}

func BenchmarkConcat(b *testing.B) {
	input := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	for i := 0; i < b.N; i++ {
		Concat(input...)
	}
}

func BenchmarkJoin(b *testing.B) {
	input := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	for i := 0; i < b.N; i++ {
		//strings.Join(input, "")
		Join("", input...)
	}
}

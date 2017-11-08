// Package base62 implements conversion of uint64 and base62(0-9A-Za-z) string.
package base62

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	base      = 62
	maxLength = 11
)

var alphabet = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

func Encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	chars := make([]byte, 0)
	for n > 0 {
		result := n / base
		remainder := n % base
		chars = append(chars, alphabet[remainder])
		n = result
	}
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func EncodeFixed(n uint64) string {
	s := Encode(n)
	if l := len(s); l < maxLength {
		s = strings.Repeat(string(alphabet[0]), maxLength-l) + s
	}
	return s
}

func Decode(str string) (uint64, error) {
	var r uint64
	for _, c := range []byte(str) {
		i := bytes.IndexByte(alphabet, c)
		if i == -1 {
			return 0, fmt.Errorf("unexpected character '%c'", c)
		}
		r = base*r + uint64(i)
	}
	return r, nil
}

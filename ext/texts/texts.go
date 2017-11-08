package texts

import (
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// Join is a fast-version of `strings.Join` function.
func Join(sep string, a ...string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0]
	case 2:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		if sep == "" {
			return a[0] + a[1]
		} else {
			return a[0] + sep + a[1]
		}
	case 3:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		if sep == "" {
			return a[0] + a[1] + a[2]
		} else {
			return a[0] + sep + a[1] + sep + a[2]
		}
	}

	var n int
	if sep != "" {
		n = len(sep) * (len(a) - 1)
	}
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	buf := make([]byte, n)
	bp := copy(buf, a[0])
	if sep == "" {
		for _, s := range a[1:] {
			bp += copy(buf[bp:], s)
		}
	} else {
		for _, s := range a[1:] {
			bp += copy(buf[bp:], sep)
			bp += copy(buf[bp:], s)
		}
	}
	return *(*string)(unsafe.Pointer(&buf))
	//return string(buf)
}

// JoinRepeat if sep = "", should use strings.Repeat instead
func JoinRepeat(s, sep string, count int) string {
	switch count {
	case 0:
		return ""
	case 1:
		return s
	case 2:
		return s + sep + s
	case 3:
		return s + sep + s + sep + s
	}

	n := (len(s)+len(sep))*(count-1) + len(s)
	buf := make([]byte, n)
	bp := copy(buf, s)
	for i := 1; i < count; i++ {
		bp += copy(buf[bp:], sep)
		bp += copy(buf[bp:], s)
	}
	return *(*string)(unsafe.Pointer(&buf))
}

func Concat(a ...string) string {
	return Join("", a...)
}

func Contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func PadLeft(s string, c byte, w int) string {
	count := w - len(s)
	if count > 0 {
		return strings.Repeat(string(c), count) + s
	}
	return s
}

func PadRight(s string, c byte, w int) string {
	count := w - len(s)
	if count > 0 {
		return s + strings.Repeat(string(c), count)
	}
	return s
}

func PadCenter(s string, c byte, w int) string {
	count := w - len(s)
	if count > 0 {
		left := count / 2
		return strings.Repeat(string(c), left) + s + strings.Repeat(string(c), count-left)
	}
	return s
}

func CutLeft(s string, w int) string {
	if len(s) > w {
		return s[len(s)-w:]
	}
	return s
}

func CutRight(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	return s
}

type NameStyle int

const (
	Pascal NameStyle = iota
	Camel
	Upper
	Lower
)

func Rename(name string, style NameStyle) string {
	words := splitName(name)
	switch style {
	case Lower:
		for i := 0; i < len(words); i++ {
			words[i] = strings.ToLower(words[i])
		}
		return strings.Join(words, "_")
	case Upper:
		for i := 0; i < len(words); i++ {
			words[i] = strings.ToUpper(words[i])
		}
		return strings.Join(words, "_")
	case Pascal:
		for i := 0; i < len(words); i++ {
			words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
		return Concat(words...)
	case Camel:
		words[0] = strings.ToLower(words[0])
		for i := 1; i < len(words); i++ {
			words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
		return Concat(words...)
	default:
		return name
	}
}

func IsUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 'A' || b > 'Z' {
			return false
		}
	}
	return true
}

func IsLower(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 'a' || b > 'z' {
			return false
		}
	}
	return true
}

func splitName(name string) (words []string) {
	var (
		start = 0
		last  byte
	)
	for i := 0; i < len(name); i++ {
		b := name[i]
		switch {
		case b == '_':
			words = append(words, name[start:i])
			start, last = i+1, 0
			continue
		case b >= 'A' && b <= 'Z': // upper char
			if last != 0 && (last < 'A' || last > 'Z') { // last is lower char
				words = append(words, name[start:i])
				start = i
			}
		default:
			if (last >= 'A' && last <= 'Z') && start != i-1 { // last is upper char
				words = append(words, name[start:i-1])
				start = i - 1
			}
		}
		last = b
	}
	words = append(words, name[start:])
	return
}

// CompareFold reports whether s and t, interpreted as UTF-8 strings,
// are equal under Unicode case-folding.
func CompareFold(s, t string) int {
	for s != "" && t != "" {
		// Extract first rune from each string.
		var sr, tr rune
		if s[0] < utf8.RuneSelf {
			sr, s = rune(s[0]), s[1:]
		} else {
			r, size := utf8.DecodeRuneInString(s)
			sr, s = r, s[size:]
		}
		if t[0] < utf8.RuneSelf {
			tr, t = rune(t[0]), t[1:]
		} else {
			r, size := utf8.DecodeRuneInString(t)
			tr, t = r, t[size:]
		}

		// If they match, keep going; if not, return false.

		// Easy case.
		if tr == sr {
			continue
		}

		// Make sr < tr to simplify what follows.
		result := 1
		if tr < sr {
			result = -result
			tr, sr = sr, tr
		}
		// Fast check for ASCII.
		if tr < utf8.RuneSelf && 'A' <= sr && sr <= 'Z' {
			// ASCII, and sr is upper case.  tr must be lower case.
			if lower := sr + 'a' - 'A'; tr == lower {
				continue
			} else if tr < lower {
				return result
			} else if tr > lower {
				return -result
			}
		}

		// General case. SimpleFold(x) returns the next equivalent rune > x
		// or wraps around to smaller values.
		r := unicode.SimpleFold(sr)
		for r != sr && r < tr {
			r = unicode.SimpleFold(r)
		}
		if r == tr {
			continue
		}
		if tr < r {
			return result
		}
		if tr > r {
			return -result
		}
	}

	// One string is empty. Are both?
	if s == "" && t != "" {
		return -1
	}
	if s != "" && t == "" {
		return 1
	}
	return 0
}

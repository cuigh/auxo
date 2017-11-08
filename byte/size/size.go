package size

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cuigh/auxo/errors"
)

// Size represents human readable bytes.
type Size uint64

const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
)

// Options represents format options
type Options struct {
	Space     bool
	Short     bool
	Precision int
}

var (
	defaultOpts = Options{
		Space:     true,
		Precision: 2,
	}
)

// Parse parses human readable bytes string to Size.
// For example, 1K/1KB(or 1 KB) will return 1024 bytes.
func Parse(value string) (s Size, err error) {
	var (
		i    int
		r    rune
		num  float64
		unit string
	)

	for i, r = range value {
		if (r < '0' || r > '9') && r != '.' {
			break
		}
	}

	if num, err = strconv.ParseFloat(value[:i], 64); err != nil {
		return
	}

	unit = strings.TrimSpace(value[i:])
	switch unit {
	case "", "B":
		s = Size(num * B)
	case "K", "KB":
		s = Size(num * KB)
	case "M", "MB":
		s = Size(num * MB)
	case "G", "GB":
		s = Size(num * GB)
	case "T", "TB":
		s = Size(num * TB)
	case "P", "PB":
		s = Size(num * PB)
	case "E", "EB":
		s = Size(num * EB)
	default:
		err = errors.New("invalid value: " + value)
	}
	return
}

// String formats Size to human readable string with specified options.
// For example, 1024 bytes will return 1KB.
func (s Size) Format(opts Options) string {
	var (
		unit  string
		value = float64(s)
	)

	switch {
	case s < KB:
		unit = "B"
	case s < MB:
		value /= KB
		unit = "KB"
	case s < GB:
		value /= MB
		unit = "MB"
	case s < TB:
		value /= GB
		unit = "GB"
	case s < PB:
		value /= TB
		unit = "TB"
	case s < EB:
		value /= PB
		unit = "PB"
	default:
		value /= EB
		unit = "EB"
	}

	ss := fmt.Sprintf("%."+strconv.Itoa(opts.Precision)+"f", value)
	i := len(ss) - 1
	for {
		r := ss[i]
		if r == '0' {
			i--
			continue
		}
		if r == '.' {
			ss = ss[:i]
		} else {
			ss = ss[:i+1]
		}
		break
	}

	if opts.Short && len(unit) > 1 {
		unit = unit[:1]
	}
	if opts.Space {
		return ss + " " + unit
	}
	return ss + unit
}

// String formats Size to human readable string with default options.
// For example, 1024 bytes will return '1 KB'.
func (s Size) String() string {
	return s.Format(defaultOpts)
}

// Unmarshal implements config.Unmarshaler interface.
func (s *Size) Unmarshal(i interface{}) (err error) {
	if value, ok := i.(string); ok {
		*s, err = Parse(value)
		return
	}
	return fmt.Errorf("unable to unmarshal %#v to Size", i)
}

func (s *Size) UnmarshalJSON(data []byte) (err error) {
	value := string(data[1 : len(data)-1]) // trim '"'
	*s, err = Parse(value)
	return
}

func (s Size) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

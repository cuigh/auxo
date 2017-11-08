package times

import (
	"bytes"
	"strconv"
	"time"

	"github.com/cuigh/auxo/ext/texts"
)

const (
	Day  = time.Hour * 24
	Week = Day * 7
)

func Milliseconds(n int32) time.Duration {
	return time.Millisecond * time.Duration(n)
}

func Seconds(n int32) time.Duration {
	return time.Second * time.Duration(n)
}

func Minutes(n int32) time.Duration {
	return time.Minute * time.Duration(n)
}

func Hours(n int32) time.Duration {
	return time.Hour * time.Duration(n)
}

func Days(n int32) time.Duration {
	return Day * time.Duration(n)
}

func Date(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

//type field struct {
//	b byte
//	bits int
//}

//type Formatter struct {
//	fields []*field
//}
//
//func Compile(layout string) *Formatter {
//	return &Formatter{}
//}
//
//// Format format time with .NET time pattern.
//func (f *Formatter) Format(t time.Time) string {
//	return ""
//}

// Format format time with .NET time pattern.
func Format(t time.Time, layout string) string {
	const (
		stateLiteral = iota
		stateEscape
		statePattern
	)
	var (
		state      = stateLiteral
		buf        = new(bytes.Buffer)
		p     byte = 0
		bits       = 0
	)
	for i := 0; i < len(layout); i++ {
		c := layout[i]

		if state == stateEscape {
			state = stateLiteral
			buf.WriteByte(c)
			continue
		}

		switch c {
		case '\\':
			if p != 0 {
				format(buf, t, p, bits)
				p = 0
			}
			state = stateEscape
		case 'y', 'M', 'd', 'H', 'h', 'm', 's', 'f', 'z':
			if state == stateLiteral {
				state = statePattern
				p = c
				bits = 1
			} else if p != c {
				format(buf, t, p, bits)
				p = c
				bits = 1
			} else {
				bits++
			}
		default:
			if state == statePattern {
				state = stateLiteral
				format(buf, t, p, bits)
				p = 0
			}
			buf.WriteByte(c)
		}
	}
	if p != 0 {
		format(buf, t, p, bits)
	}
	return buf.String()
}

func format(buf *bytes.Buffer, t time.Time, b byte, bits int) {
	var s string
	switch b {
	case 'y':
		s = texts.CutLeft(strconv.Itoa(t.Year()), bits)
	case 'M':
		if bits < 3 {
			s = texts.PadLeft(strconv.Itoa(int(t.Month())), '0', bits)
		} else if bits == 3 {
			s = texts.CutLeft(t.Month().String(), bits)
		} else {
			s = t.Month().String()
		}
	case 'd':
		if bits < 3 {
			s = texts.PadLeft(strconv.Itoa(t.Day()), '0', bits)
		} else if bits == 3 {
			s = texts.CutLeft(t.Weekday().String(), bits)
		} else {
			s = t.Weekday().String()
		}
	case 'h':
		hr := t.Hour() % 12
		if hr == 0 {
			hr = 12
		}
		s = texts.PadLeft(strconv.Itoa(hr), '0', bits)
	case 'H':
		s = texts.PadLeft(strconv.Itoa(t.Hour()), '0', bits)
	case 'm':
		s = texts.PadLeft(strconv.Itoa(t.Minute()), '0', bits)
	case 's':
		s = texts.PadLeft(strconv.Itoa(t.Second()), '0', bits)
	case 'f':
		ns := t.Nanosecond()
		s = texts.CutRight(texts.PadLeft(strconv.Itoa(ns), '0', 9), bits)
	case 'z':
		_, offset := t.Zone()
		minutes := offset / 60
		hour := minutes / 60
		s = strconv.Itoa(hour)
		if bits > 1 {
			s = texts.PadLeft(s, '0', 2)
		}
		if offset > 0 {
			s = "+" + s
		}
		if bits > 2 {
			s = s + ":" + texts.PadLeft(strconv.Itoa(minutes%60), '0', 2)
		}
	}
	buf.WriteString(s)
}

package log

import (
	"bytes"
	"fmt"
	"strings"
)

type Segment struct {
	Type string
	Name string
	Args []string
}

type Layout interface {
	Parse(s string) ([]Segment, error)
}

type baseLayout struct {
}

func (baseLayout) create(s string) (field Segment) {
	array := strings.SplitN(s, ":", 2)
	if len(array) > 1 {
		field.Args = strings.Split(strings.TrimSpace(array[1]), "|")
	}

	pair := strings.SplitN(array[0], "->", 2)
	field.Type = strings.TrimSpace(pair[0])
	if len(pair) == 1 {
		field.Name = field.Type
	} else {
		field.Name = strings.TrimSpace(pair[1])
	}
	return
}

type JSONLayout struct {
	baseLayout
}

func (l JSONLayout) Parse(s string) ([]Segment, error) {
	// {level->lvl: a=b},{time->t:2016-01-02},{msg->msg},{file->f: s},{text->abc: test}
	const (
		stateSeparator = 0 // comma
		stateLeftFlag  = 1 // {
		stateField     = 2 // field
		stateRightFlag = 3 // }
	)

	var (
		fields []Segment
		buf    bytes.Buffer
		state  = stateSeparator
	)

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if state == stateField {
				buf.WriteByte(s[i])
			} else if state == stateSeparator {
				state = stateLeftFlag
			} else {
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		case '}':
			if state == stateRightFlag {
				buf.WriteByte(s[i])
			} else if state == stateField {
				state = stateRightFlag
			} else {
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		case ',':
			if state == stateRightFlag {
				field := l.create(buf.String())
				fields = append(fields, field)
				buf.Reset()
				state = stateSeparator
			} else if state == stateField {
				buf.WriteByte(s[i])
			} else {
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		default:
			switch state {
			case stateLeftFlag:
				state = stateField
				fallthrough
			case stateField:
				buf.WriteByte(s[i])
			default:
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		}
	}

	if buf.Len() > 0 {
		if state == stateRightFlag {
			field := l.create(buf.String())
			fields = append(fields, field)
		} else {
			return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, len(s)-1)
		}
	}

	return fields, nil
}

type TextLayout struct {
	baseLayout
}

func (l TextLayout) Parse(s string) ([]Segment, error) {
	const (
		stateString    = 0 // text
		stateLeftFlag  = 1 // {
		stateField     = 2 // field
		stateRightFlag = 3 // }
	)

	var (
		segments []Segment
		buf      bytes.Buffer
		state    = stateString
	)

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if state == stateLeftFlag {
				buf.WriteByte(s[i])
			} else if state == stateString {
				state = stateLeftFlag
			} else if state == stateRightFlag {
				segment := l.create(buf.String())
				segments = append(segments, segment)
				buf.Reset()
				state = stateLeftFlag
			} else {
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		case '}':
			if state == stateRightFlag {
				buf.WriteByte(s[i])
			} else if state == stateField {
				state = stateRightFlag
			} else {
				return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, i)
			}
		default:
			if state == stateLeftFlag {
				if buf.Len() > 0 {
					segment := Segment{Type: "text", Args: []string{buf.String()}}
					segments = append(segments, segment)
					buf.Reset()
				}
				state = stateField
			} else if state == stateRightFlag {
				segment := l.create(buf.String())
				segments = append(segments, segment)
				buf.Reset()
				state = stateString
			}
			buf.WriteByte(s[i])
		}
	}

	if buf.Len() > 0 {
		if state == stateRightFlag {
			segment := l.create(buf.String())
			segments = append(segments, segment)
		} else if state == stateString {
			segment := Segment{Type: "text", Args: []string{buf.String()}}
			segments = append(segments, segment)
		} else {
			return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, len(s)-1)
		}
	}

	return segments, nil
}

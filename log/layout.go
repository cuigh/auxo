package log

import (
	"bytes"
	"fmt"
	"strings"
)

type Layout interface {
	Parse(s string) ([]Field, error)
}

type jsonLayout struct {
}

func (jsonLayout) Parse(s string) ([]Field, error) {
	const (
		stateSeparator = 0 // comma
		stateLeftFlag  = 1 // {
		stateField     = 2 // field
		stateRightFlag = 3 // }
	)

	var (
		fields []Field
		buf    bytes.Buffer
		state  = stateSeparator
	)

	creator := func(s string) (Field, error) {
		array := strings.SplitN(s, ":", 2)
		var args string
		if len(array) == 1 {
			args = ""
		} else {
			args = strings.TrimSpace(array[1])
		}
		field, err := newField(strings.TrimSpace(array[0]), args)
		if err != nil {
			return nil, err
		}
		return field, nil
	}

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
				field, err := creator(buf.String())
				if err != nil {
					return nil, err
				}

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
			field, err := creator(buf.String())
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		} else {
			return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, len(s)-1)
		}
	}

	return fields, nil
}

type textLayout struct {
}

func (textLayout) Parse(s string) ([]Field, error) {
	const (
		stateString    = 0 // text
		stateLeftFlag  = 1 // {
		stateField     = 2 // field
		stateRightFlag = 3 // }
	)

	var (
		fields []Field
		buf    bytes.Buffer
		state  = stateString
	)

	creator := func(s string) (Field, error) {
		array := strings.SplitN(s, ":", 2)
		var args string
		if len(array) == 1 {
			args = ""
		} else {
			args = array[1]
		}
		field, err := newField(array[0], args)
		if err != nil {
			return nil, err
		}
		return field, nil
	}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if state == stateLeftFlag {
				buf.WriteByte(s[i])
			} else if state == stateString {
				state = stateLeftFlag
			} else if state == stateRightFlag {
				field, err := creator(buf.String())
				if err != nil {
					return nil, err
				}
				fields = append(fields, field)
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
					fields = append(fields, newStringField("", buf.String()))
					buf.Reset()
				}
				state = stateField
			} else if state == stateRightFlag {
				field, err := creator(buf.String())
				if err != nil {
					return nil, err
				}
				fields = append(fields, field)
				buf.Reset()
				state = stateString
			}
			buf.WriteByte(s[i])
		}
	}

	if buf.Len() > 0 {
		if state == stateRightFlag {
			field, err := creator(buf.String())
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		} else if state == stateString {
			fields = append(fields, newStringField("", buf.String()))
		} else {
			return nil, fmt.Errorf("invalid layout: %s(pos: %d)", s, len(s)-1)
		}
	}

	return fields, nil
}

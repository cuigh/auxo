package log

import (
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"unsafe"
)

const (
	fieldLevel   = "L"
	fieldMessage = "M"
	fieldTime    = "T"
	fieldFile    = "F"
	fieldNewline = "N"
	//fieldContext = "C" // 上下文 ID
)

type field interface {
	Name() string
	Write(w io.Writer, e *entry) error
	Value(e *entry) interface{}
}

func newField(seg *Segment) (field, error) {
	switch seg.Type {
	case fieldLevel, "level":
		return newLevelField(seg.Name, seg.Args...), nil
	case fieldMessage, "msg":
		return newMessageField(seg.Name), nil
	case fieldTime, "time":
		return newTimeField(seg.Name, seg.Args...), nil
	case fieldFile, "file":
		return newFileField(seg.Name, seg.Args...), nil
	case fieldNewline, "newline":
		return newStringField(seg.Name, "\n"), nil
	case "text":
		return newStringField(seg.Name, seg.Args[0]), nil
	default:
		return nil, errors.New("invalid field: " + seg.Type)
	}
}

/********** baseField **********/

type baseField string

func (f baseField) Name() string {
	return string(f)
}

func (f baseField) writeString(w io.Writer, s string) (err error) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	b := *(*[]byte)(unsafe.Pointer(&h))
	_, err = w.Write(b)
	return
}

/********** stringField **********/

type stringField struct {
	baseField
	s string
}

func newStringField(name string, s string) *stringField {
	f := &stringField{
		baseField: baseField(name),
		s:         s,
	}
	return f
}

func (f stringField) Value(e *entry) interface{} {
	return f.s
}

func (f stringField) Write(w io.Writer, e *entry) (err error) {
	return f.writeString(w, f.s)
}

/********** levelField **********/

type levelField struct {
	baseField
	texts [7]string
}

func newLevelField(name string, args ...string) *levelField {
	f := &levelField{
		baseField: baseField(name),
	}
	if len(args) == 0 || args[0] == "S" || args[0] == "s" || args[0] == "short" {
		f.texts = levelShortNames
	} else {
		f.texts = levelLongNames
	}
	return f
}

func (f levelField) Value(e *entry) interface{} {
	return f.texts[e.lvl]
}

func (f levelField) Write(w io.Writer, e *entry) (err error) {
	s := f.texts[e.lvl]
	return f.writeString(w, s)
}

/********** timeField **********/

type timeField struct {
	baseField
	layout string
}

func newTimeField(name string, args ...string) *timeField {
	f := &timeField{
		baseField: baseField(name),
	}
	if len(args) == 0 {
		f.layout = "2006-01-02 15:04:05.000"
	} else {
		f.layout = args[0]
	}
	return f
}

func (f timeField) Value(e *entry) interface{} {
	return e.time.Format(f.layout)
}

func (f timeField) Write(w io.Writer, e *entry) (err error) {
	s := e.time.Format(f.layout)
	return f.writeString(w, s)
}

/********** fileField **********/

type fileField struct {
	baseField
	full bool
	skip int
}

func newFileField(name string, args ...string) *fileField {
	// todo: support set min level
	//{F:S|L=E}
	f := &fileField{
		baseField: baseField(name),
		skip:      8,
		full:      len(args) == 0 || args[0] == "F" || args[0] == "full",
	}
	return f
}

func (f fileField) Value(e *entry) interface{} {
	return f.file()
}

func (f fileField) Write(w io.Writer, e *entry) (err error) {
	return f.writeString(w, f.file())
}

func (f fileField) file() string {
	_, file, line, ok := runtime.Caller(f.skip)
	if ok {
		if !f.full {
			file = filepath.Base(file)
		}
		return file + ":" + strconv.Itoa(line)
	} else {
		return "?file?"
	}
}

/********** messageField **********/

type messageField struct {
	baseField
}

func newMessageField(name string) field {
	return &messageField{
		baseField: baseField(name),
	}
}

func (f messageField) Value(e *entry) interface{} {
	return e.msg
}

func (f messageField) Write(w io.Writer, e *entry) (err error) {
	return f.writeString(w, e.msg)
}

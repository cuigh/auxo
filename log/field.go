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

type Field interface {
	Name() string
	Write(w io.Writer, r *Row) error
	Value(r *Row) interface{}
}

func newField(seg *Segment) (Field, error) {
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

func (f stringField) Value(r *Row) interface{} {
	return f.s
}

func (f stringField) Write(w io.Writer, r *Row) (err error) {
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

func (f levelField) Value(r *Row) interface{} {
	return f.texts[r.lvl]
}

func (f levelField) Write(w io.Writer, r *Row) (err error) {
	s := f.texts[r.lvl]
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

func (f timeField) Value(r *Row) interface{} {
	return r.time.Format(f.layout)
}

func (f timeField) Write(w io.Writer, r *Row) (err error) {
	s := r.time.Format(f.layout)
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
		skip:      7,
		full:      len(args) == 0 || args[0] == "F" || args[0] == "full",
	}
	return f
}

func (f fileField) Value(r *Row) interface{} {
	return f.file()
}

func (f fileField) Write(w io.Writer, r *Row) (err error) {
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

func newMessageField(name string) Field {
	return &messageField{
		baseField: baseField(name),
	}
}

func (f messageField) Value(r *Row) interface{} {
	return r.msg
}

func (f messageField) Write(w io.Writer, r *Row) (err error) {
	return f.writeString(w, r.msg)
}

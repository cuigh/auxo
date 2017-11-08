package log

import (
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

func newField(name string, option string) (Field, error) {
	var (
		ft string
		fn string
	)

	pair := strings.SplitN(name, "->", 2)
	ft = strings.TrimSpace(pair[0])
	if len(pair) > 1 {
		fn = strings.TrimSpace(pair[1])
	} else {
		fn = ft
	}

	var args []string = nil
	if option != "" {
		args = strings.Split(option, "|")
	}

	switch ft {
	case fieldLevel, "level":
		return newLevelField(fn, args...), nil
	case fieldMessage, "msg":
		return newMessageField(fn), nil
	case fieldTime, "time":
		return newTimeField(fn, args...), nil
	case fieldFile, "file":
		return newFileField(fn, args...), nil
	case fieldNewline:
		return newStringField(fn, "\n"), nil
	case "text":
		return newStringField(fn, args[0]), nil
	default:
		return nil, errors.New("invalid field: " + name)
	}
}

// {level->lvl: a=b},{time->t:2016-01-02},{msg->msg},{file->f: s},{text->abc: test}
//func newField1(name string, strArgs string) (Field, error) {
//	var (
//		alias string
//	)
//
//	pair := strings.SplitN(name, "->", 2)
//	if len(pair) > 1 {
//		alias = strings.TrimSpace(pair[1])
//	}
//
//	var args []string = nil
//	if strArgs != "" {
//		args = strings.Split(strArgs, "|")
//	}
//
//	switch name {
//	case fieldLevel:
//		return newLevelField(args...), nil
//	case fieldMessage:
//		return newMessageField(args...), nil
//	case fieldTime:
//		return newTimeField(args...), nil
//	case fieldFile:
//		return newFileField(args...), nil
//	case fieldNewline:
//		return stringField("\n"), nil
//	default:
//		return nil, errors.New("invalid field: " + name)
//	}
//}

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
	if len(args) < 1 || args[0] == "S" {
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
	// todo: 增加最小 level 控制
	//{F:S|L=E}
	f := &fileField{
		baseField: baseField(name),
		skip:      7,
	}
	if len(args) == 0 {
		f.full = true
	} else {
		f.full = args[0] == "F"
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

package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/log/console"
	"github.com/cuigh/auxo/log/file"
)

var writerBuilders = WriterBuilders{
	"": func(options data.Map) (io.Writer, error) {
		return os.Open(os.DevNull)
	},
	"console": func(options data.Map) (io.Writer, error) {
		return console.New(options)
	},
	"file": func(options data.Map) (io.Writer, error) {
		return file.New(options)
	},
}

type WriterBuilder func(options data.Map) (io.Writer, error)

type WriterBuilders map[string]WriterBuilder

func (bs WriterBuilders) Build(name string, options data.Map) (io.Writer, error) {
	b, ok := bs[name]
	if !ok {
		return nil, errors.New("builder not found: " + name)
	}

	return b(options)
}

func RegisterWriter(name string, b WriterBuilder) {
	writerBuilders[name] = b
}

type Writer struct {
	name    string
	format  string
	out     io.Writer
	fields  []field
	options data.Map
	write   func(fields []field, buf *bytes.Buffer, e *entry) error
}

func (w *Writer) Name() string {
	return w.name
}

func (w *Writer) Output() io.Writer {
	return w.out
}

func (w *Writer) Write(e *entry) (err error) {
	e.buf.Reset()
	if err = w.write(w.fields, e.buf, e); err == nil {
		_, err = w.out.Write(e.buf.Bytes())
	}
	return
}

func newWriter(name, typeName, format, layout string, options data.Map) (*Writer, error) {
	out, err := writerBuilders.Build(typeName, options)
	if err != nil {
		return nil, err
	}

	var (
		parser Layout
		write  func(fields []field, buf *bytes.Buffer, e *entry) error
	)

	if format == "json" {
		parser = JSONLayout{}
		write = func(fields []field, buf *bytes.Buffer, e *entry) (err error) {
			m := e.fields
			if m == nil {
				m = map[string]interface{}{}
			}
			for _, f := range fields {
				m[f.Name()] = f.Value(e)
			}
			return json.NewEncoder(buf).Encode(m)
		}
	} else {
		parser = TextLayout{}
		write = func(fields []field, buf *bytes.Buffer, e *entry) (err error) {
			for _, f := range fields {
				if err = f.Write(buf, e); err != nil {
					return
				}
			}
			return
		}
	}

	segments, err := parser.Parse(layout)
	if err != nil {
		return nil, err
	}

	fields := make([]field, len(segments))
	for i, seg := range segments {
		fields[i], err = newField(&seg)
		if err != nil {
			return nil, err
		}
	}

	return &Writer{
		name:    name,
		format:  format,
		out:     out,
		fields:  fields,
		options: options,
		write:   write,
	}, nil
}

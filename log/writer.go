package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

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

type Writer struct {
	name    string
	format  string
	out     io.Writer
	fields  []Field
	options data.Map
	bufs    sync.Pool
	write   func(fields []Field, buf *bytes.Buffer, r *Row) error
}

func RegisterWriter(name string, b WriterBuilder) {
	writerBuilders[name] = b
}

func (w *Writer) Name() string {
	return w.name
}

func (w *Writer) Output() io.Writer {
	return w.out
}

func (w *Writer) Write(r *Row) (err error) {
	buf := w.bufs.Get().(*bytes.Buffer)
	buf.Reset()
	if err = w.write(w.fields, buf, r); err == nil {
		_, err = w.out.Write(buf.Bytes())
	}
	w.bufs.Put(buf)
	return
}

func newWriter(name, typeName, format, layout string, options data.Map) (*Writer, error) {
	out, err := writerBuilders.Build(typeName, options)
	if err != nil {
		return nil, err
	}

	var (
		parser Layout
		write  func(fields []Field, buf *bytes.Buffer, r *Row) error
	)

	if format == "json" {
		parser = jsonLayout{}
		write = func(fields []Field, buf *bytes.Buffer, r *Row) (err error) {
			m := map[string]interface{}{}
			for _, f := range fields {
				m[f.Name()] = f.Value(r)
			}
			return json.NewEncoder(buf).Encode(m)
		}
	} else {
		parser = textLayout{}
		write = func(fields []Field, buf *bytes.Buffer, r *Row) (err error) {
			for _, f := range fields {
				if err = f.Write(buf, r); err != nil {
					return
				}
			}
			return
		}
	}

	fields, err := parser.Parse(layout)
	if err != nil {
		return nil, err
	}

	return &Writer{
		name:    name,
		format:  format,
		out:     out,
		fields:  fields,
		options: options,
		write:   write,
		bufs: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}, nil
}

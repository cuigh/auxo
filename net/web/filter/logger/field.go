package logger

import (
	"time"

	"strconv"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/net/web"
)

type field interface {
	Text(c web.Context, b *texts.Builder, start time.Time)
	JSON(c web.Context, m data.Map, start time.Time)
}

type timeField struct {
	layout string
}

func (f *timeField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(start.Format(f.layout))
}

func (f *timeField) JSON(c web.Context, m data.Map, start time.Time) {
	m["time"] = start.Format(f.layout)
}

type statusField struct{}

func (statusField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(strconv.Itoa(c.Response().Status()))
}

func (statusField) JSON(c web.Context, m data.Map, start time.Time) {
	m["status"] = c.Response().Status()
}

type latencyField struct {
	//unit string // ns/ms/s
}

func (latencyField) Text(c web.Context, b *texts.Builder, start time.Time) {
	latency := time.Since(start)
	b.Append(latency.String())
}

func (latencyField) JSON(c web.Context, m data.Map, start time.Time) {
	m["latency"] = time.Since(start)
}

type ipField struct{}

func (ipField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(c.RealIP())
}

func (ipField) JSON(c web.Context, m data.Map, start time.Time) {
	m["ip"] = c.RealIP()
}

type methodField struct{}

func (methodField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(c.Request().Method)
}

func (methodField) JSON(c web.Context, m data.Map, start time.Time) {
	m["method"] = c.Request().Method
}

type pathField struct{}

func (pathField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(c.Request().RequestURI)
}

func (pathField) JSON(c web.Context, m data.Map, start time.Time) {
	m["path"] = c.Request().RequestURI
}

type headerField struct {
	name string
}

func (f *headerField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.Append(c.Header(f.name))
}

func (f *headerField) JSON(c web.Context, m data.Map, start time.Time) {
	m[f.name] = c.Header(f.name)
}

func parseFields(layout string) []field {
	// - time
	// - ip (remote_ip)
	// - uri
	// - host
	// - method
	// - path
	// - referer
	// - user_agent
	// - status
	// - latency (In nanoseconds)
	// - header:<NAME>
	// - query:<NAME>
	// - form:<NAME>
	// e.g. {time: 2006/01/02 - 15:04:05},{status},{latency},{ip},{method},{path},{header: Cookie},{form: name}
	// todo:
	return []field{
		&timeField{layout: "2006/01/02 - 15:04:05"},
		&statusField{},
		&latencyField{},
		&ipField{},
		&methodField{},
		&pathField{},
	}
}

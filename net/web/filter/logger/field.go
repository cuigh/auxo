package logger

import (
	"errors"
	"strconv"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web"
)

type field interface {
	Text(c web.Context, b *texts.Builder, start time.Time)
	JSON(c web.Context, m data.Map, start time.Time)
}

type textField struct {
	name string
	text string
}

func (f *textField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(f.text)
}

func (f *textField) JSON(c web.Context, m data.Map, start time.Time) {
	m[f.name] = f.text
}

type timeField struct {
	layout string
}

func (f *timeField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(start.Format(f.layout))
}

func (f *timeField) JSON(c web.Context, m data.Map, start time.Time) {
	m["time"] = start.Format(f.layout)
}

type statusField struct{}

func (statusField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(strconv.Itoa(c.Response().Status()))
}

func (statusField) JSON(c web.Context, m data.Map, start time.Time) {
	m["status"] = c.Response().Status()
}

type latencyField struct {
	//unit string // ns/ms/s
}

func (latencyField) Text(c web.Context, b *texts.Builder, start time.Time) {
	latency := time.Since(start)
	b.WriteString(latency.String())
}

func (latencyField) JSON(c web.Context, m data.Map, start time.Time) {
	m["latency"] = time.Since(start)
}

type ipField struct{}

func (ipField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.RealIP())
}

func (ipField) JSON(c web.Context, m data.Map, start time.Time) {
	m["ip"] = c.RealIP()
}

type methodField struct{}

func (methodField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Request().Method)
}

func (methodField) JSON(c web.Context, m data.Map, start time.Time) {
	m["method"] = c.Request().Method
}

type hostField struct{}

func (hostField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Request().Host)
}

func (hostField) JSON(c web.Context, m data.Map, start time.Time) {
	m["host"] = c.Request().Host
}

type pathField struct{}

func (pathField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Request().RequestURI)
}

func (pathField) JSON(c web.Context, m data.Map, start time.Time) {
	m["path"] = c.Request().RequestURI
}

type refererField struct{}

func (refererField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Request().Referer())
}

func (refererField) JSON(c web.Context, m data.Map, start time.Time) {
	m["referer"] = c.Request().Referer
}

type userAgentField struct{}

func (userAgentField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Request().UserAgent())
}

func (userAgentField) JSON(c web.Context, m data.Map, start time.Time) {
	m["user_agent"] = c.Request().UserAgent
}

type headerField struct {
	name string
}

func (f *headerField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Header(f.name))
}

func (f *headerField) JSON(c web.Context, m data.Map, start time.Time) {
	m[f.name] = c.Header(f.name)
}

type queryField struct {
	name string
}

func (f *queryField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Query(f.name))
}

func (f *queryField) JSON(c web.Context, m data.Map, start time.Time) {
	m[f.name] = c.Query(f.name)
}

type formField struct {
	name string
}

func (f *formField) Text(c web.Context, b *texts.Builder, start time.Time) {
	b.WriteString(c.Form(f.name))
}

func (f *formField) JSON(c web.Context, m data.Map, start time.Time) {
	m[f.name] = c.Form(f.name)
}

func createFields(layout string, json bool) []field {
	// - time
	// - ip (remote_ip)
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
	var (
		segments []log.Segment
		err      error
	)

	if json {
		segments, err = log.JSONLayout{}.Parse(layout)
	} else {
		segments, err = log.TextLayout{}.Parse(layout)
	}
	if err != nil {
		panic(err)
	}

	fields := make([]field, len(segments))
	for i, seg := range segments {
		var f field
		switch seg.Type {
		case "text":
			f = &textField{name: seg.Name, text: seg.Args[0]}
		case "time":
			f = &timeField{layout: seg.Args[0]}
		case "status":
			f = statusField{}
		case "latency":
			f = latencyField{}
		case "ip":
			f = ipField{}
		case "method":
			f = methodField{}
		case "host":
			f = hostField{}
		case "path":
			f = pathField{}
		case "referer":
			f = refererField{}
		case "user_agent":
			f = userAgentField{}
		case "header":
			f = &headerField{name: seg.Args[0]}
		case "query":
			f = &queryField{name: seg.Args[0]}
		case "form":
			f = &formField{name: seg.Args[0]}
		default:
			panic(errors.New("invalid field: " + seg.Type))
		}
		fields[i] = f
	}
	return fields
}

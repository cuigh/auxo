package logger

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web"
)

const (
	PkgName           = "auxo.net.web.filter.logger"
	defaultJSONLayout = "{time: 2006/01/02 - 15:04:05},{status},{latency},{ip},{method},{path}"
	defaultTextLayout = "{time: 2006/01/02 - 15:04:05} {status} {latency} {ip} {method} {path}"
)

type Option func(*Logger)

type Logger struct {
	layout string
	format string
	logger log.Logger
	fields []field
}

func Layout(layout string) Option {
	return func(l *Logger) {
		if layout != "" {
			l.layout = layout
		}
	}
}

func Format(format string) Option {
	return func(l *Logger) {
		if format != "" {
			l.format = strings.ToLower(format)
		}
	}
}

func New(opts ...Option) *Logger {
	l := &Logger{
		format: "text",
	}
	for _, opt := range opts {
		opt(l)
	}
	if l.layout == "" {
		if l.format == "json" {
			l.layout = defaultJSONLayout
		} else {
			l.layout = defaultTextLayout
		}
	}
	return l
}

func (l *Logger) Apply(next web.HandlerFunc) web.HandlerFunc {
	isJSON := strings.EqualFold(l.format, "json")
	l.fields = createFields(l.layout, isJSON)
	l.logger = log.Get(PkgName)

	return func(ctx web.Context) (err error) {
		start := time.Now()
		if err = next(ctx); err != nil {
			ctx.Error(err)
		}

		if isJSON {
			l.json(ctx, start)
		} else {
			l.text(ctx, start)
		}
		return
	}
}

func (l *Logger) text(ctx web.Context, start time.Time) {
	b := texts.GetBuilder()
	for _, f := range l.fields {
		f.Text(ctx, b, start)
	}
	b.WriteByte('\n')
	l.logger.Write(b.Bytes())
	texts.PutBuilder(b)
}

func (l *Logger) json(ctx web.Context, start time.Time) {
	m := data.Map{}
	for _, f := range l.fields {
		f.JSON(ctx, m, start)
	}
	b := texts.GetBuilder()
	err := json.NewEncoder(b).Encode(m)
	if err == nil {
		l.logger.Write(b.Bytes())
	}
	texts.PutBuilder(b)
}

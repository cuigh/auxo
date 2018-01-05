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

const PkgName = "auxo.net.web.filter.logger"

type Logger struct {
	Layout string
	Format string
	logger *log.Logger
	fields []field
}

func (l *Logger) Apply(next web.HandlerFunc) web.HandlerFunc {
	l.logger = log.Get(PkgName)
	isJSON := strings.EqualFold(l.Format, "json")
	if l.Layout == "" {
		l.Layout = "{time: 2006/01/02 - 15:04:05},{status},{latency},{ip},{method},{path}"
	}
	l.fields = parseFields(l.Layout)

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
	for i, f := range l.fields {
		if i > 0 {
			b.AppendByte(' ')
		}
		f.Text(ctx, b, start)
	}
	b.AppendByte('\n')
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

package trace

import (
	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/net/web"
	"github.com/opentracing/opentracing-go/ext"
)

type Trace struct {
	*trace.Tracer
	Component string
}

func (t *Trace) Apply(next web.HandlerFunc) web.HandlerFunc {
	tracer := t.Tracer
	if tracer == nil {
		tracer = trace.GetTracer()
	}

	component := t.Component
	if component == "" {
		component = "web"
	}

	return func(c web.Context) error {
		r := c.Request()
		span := tracer.StartServer(c.Handler().Name(), trace.HTTPHeaders, trace.HTTPHeadersCarrier(r.Header))
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())
		ext.Component.Set(span, component)
		c.SetRequest(r.WithContext(trace.ContextWithSpan(r.Context(), span)))
		defer func() {
			ext.HTTPStatusCode.Set(span, uint16(c.Response().Status()))
			span.Finish()
		}()

		return next(c)
	}
}

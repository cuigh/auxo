package trace

import (
	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/rpc"
	"github.com/opentracing/opentracing-go/ext"
)

type rpcLabels data.Options

// Set implements opentracing.TextMapWriter interface.
func (c *rpcLabels) Set(key, val string) {
	*c = append(*c, data.Option{Name: key, Value: val})
}

// ForeachKey implements opentracing.TextMapCarrier interface.
func (c rpcLabels) ForeachKey(handler func(key, val string) error) error {
	for _, opt := range c {
		if err := handler(opt.Name, opt.Value); err != nil {
			return err
		}
	}
	return nil
}

func Server() rpc.SFilter {
	return func(next rpc.SHandler) rpc.SHandler {
		const component = "rpc/server"
		tracer := trace.GetTracer()

		return func(c rpc.Context) (r interface{}, err error) {
			span := tracer.StartServer(c.Action().Name(), trace.TextMap, rpcLabels(c.Request().Head.Labels))
			ext.Component.Set(span, component)
			c.SetContext(trace.ContextWithSpan(c.Context(), span))
			defer func() {
				span.SetTag("rpc.status_code", rpc.StatusOf(err))
				span.Finish()
			}()

			r, err = next(c)
			return
		}
	}
}

func Client() rpc.CFilter {
	return func(next rpc.CHandler) rpc.CHandler {
		const component = "rpc/client"
		tracer := trace.GetTracer()

		return func(c *rpc.Call) (err error) {
			req := c.Request()
			span := tracer.StartChildFromContext(c.Context(), req.Head.Service+"."+req.Head.Method)
			ext.SpanKindRPCClient.Set(span)
			ext.Component.Set(span, component)
			tracer.Inject(span.Context(), trace.TextMap, (*rpcLabels)(&req.Head.Labels))
			//c.SetContext(trace.ContextWithSpan(c.Context(), span))
			defer func() {
				span.SetTag("rpc.status_code", rpc.StatusOf(err))
				span.Finish()
			}()

			err = next(c)
			return
		}
	}
}

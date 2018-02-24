package trace

import (
	"context"
	"io"

	"github.com/cuigh/auxo/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	PkgName = "auxo.apm.trace"

	Binary      = opentracing.Binary
	TextMap     = opentracing.TextMap
	HTTPHeaders = opentracing.HTTPHeaders
)

var (
	global *Tracer
)

type Span = opentracing.Span
type StartSpanOption = opentracing.StartSpanOption
type HTTPHeadersCarrier = opentracing.HTTPHeadersCarrier
type TextMapCarrier = opentracing.TextMapCarrier

func SetTracer(t opentracing.Tracer) {
	opentracing.SetGlobalTracer(t)
	global = NewTracer(t)
}

func GetTracer() *Tracer {
	return global
}

func StartSpan(operation string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return opentracing.StartSpan(operation, opts...)
}

func StartChild(parent opentracing.Span, operation string) (span opentracing.Span) {
	if parent == nil {
		span = opentracing.StartSpan(operation)
	} else {
		span = opentracing.StartSpan(operation, opentracing.ChildOf(parent.Context()))
	}
	return
}

// StartChildFromContext starts and returns a Span with `operation`, using
// any Span found within `ctx` as a ChildOfRef. If no such parent could be
// found, StartChildFromContext creates a root (parentless) Span.
//
// Example usage:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp := trace.StartChildFromContext(ctx, "SomeFunction")
//        defer sp.Finish()
//        ...
//    }
func StartChildFromContext(ctx context.Context, operation string, opts ...opentracing.StartSpanOption) (span opentracing.Span) {
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
		span = opentracing.StartSpan(operation, opts...)
	} else {
		span = opentracing.StartSpan(operation, opts...)
	}
	return
}

func StartFollow(follow opentracing.Span, operation string) (span opentracing.Span) {
	if follow == nil {
		span = opentracing.StartSpan(operation)
	} else {
		span = opentracing.StartSpan(operation, opentracing.FollowsFrom(follow.Context()))
	}
	return
}

func Extract(format, carrier interface{}) opentracing.SpanContext {
	return GetTracer().Extract(format, carrier)
}

func Inject(sc opentracing.SpanContext, format, carrier interface{}) {
	GetTracer().Inject(sc, format, carrier)
}

func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// StartServer extracts SpanContext from carrier, then starts and returns a server Span with `operation`,
// using any Span found within `SpanContext` as a ChildOfRef. If no such parent could be
// found, StartChildFromContext creates a root (parentless) Span.
//
// Example usage:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp := trace.StartServer("SomeFunction", trace.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
//        defer sp.Finish()
//        ...
//    }
func StartServer(operation string, format, carrier interface{}) opentracing.Span {
	return GetTracer().StartServer(operation, format, carrier)
}

//func RequestWithSpan(r *http.Request, span opentracing.Span) *http.Request {
//	return r.WithContext(ContextWithSpan(r.Context(), span))
//}
//
//func SpanFromRequest(r *http.Request) opentracing.Span {
//	return SpanFromContext(r.Context())
//}

//func SetSpanKind(span opentracing.Span, enum ext.SpanKindEnum) {
//	span.SetTag("span.kind", enum)
//	ext.SpanKindRPCClient.Set(span)
//}

type Tracer struct {
	opentracing.Tracer
	closer io.Closer
	logger log.Logger
}

func NewTracer(tracer opentracing.Tracer) *Tracer {
	if tracer == nil {
		panic("nil tracer")
	}
	return &Tracer{
		Tracer: tracer,
		logger: log.Get(PkgName),
	}
}

func (t *Tracer) StartChild(parent opentracing.Span, operation string) opentracing.Span {
	var span opentracing.Span
	if parent == nil {
		span = t.StartSpan(operation)
	} else {
		span = t.StartSpan(operation, opentracing.ChildOf(parent.Context()))
	}
	return span
}

func (t *Tracer) StartFollow(follow opentracing.Span, operation string) opentracing.Span {
	var span opentracing.Span
	if follow == nil {
		span = t.StartSpan(operation)
	} else {
		span = t.StartSpan(operation, opentracing.FollowsFrom(follow.Context()))
	}
	return span
}

// StartChildFromContext starts and returns a Span with `operation`, using
// any Span found within `ctx` as a ChildOfRef. If no such parent could be
// found, StartChildFromContext creates a root (parentless) Span.
//
// Example usage:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp := trace.StartChildFromContext(ctx, "SomeFunction")
//        defer sp.Finish()
//        ...
//    }
func (t *Tracer) StartChildFromContext(ctx context.Context, operation string, opts ...opentracing.StartSpanOption) opentracing.Span {
	var span opentracing.Span
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
		span = t.StartSpan(operation, opts...)
	} else {
		span = t.StartSpan(operation, opts...)
	}
	return span
}

// StartServer extracts SpanContext from carrier, then starts and returns a server Span with `operation`,
// using any Span found within `SpanContext` as a ChildOfRef. If no such parent could be
// found, StartChildFromContext creates a root (parentless) Span.
//
// Example usage:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp := trace.StartServer("SomeFunction", trace.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
//        defer sp.Finish()
//        ...
//    }
func (t *Tracer) StartServer(operation string, format, carrier interface{}) opentracing.Span {
	ctx := t.Extract(format, carrier)
	return t.StartSpan(operation, ext.RPCServerOption(ctx))
}

// Extract() returns a SpanContext instance given `format` and `carrier`.
// If there was simply no SpanContext to extract or errors occurred, Extract() returns nil.
func (t *Tracer) Extract(format, carrier interface{}) opentracing.SpanContext {
	sc, err := t.Tracer.Extract(format, carrier)
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		t.logger.Debug("trace > ", err)
	}
	return sc
}

func (t *Tracer) Inject(sc opentracing.SpanContext, format, carrier interface{}) {
	err := t.Tracer.Inject(sc, format, carrier)
	if err != nil {
		t.logger.Debug("trace > ", err)
	}
}

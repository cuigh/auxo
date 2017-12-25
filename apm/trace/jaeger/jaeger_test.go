package jaeger_test

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/apm/trace/jaeger"
	"github.com/opentracing/opentracing-go"
)

func TestTrace(t *testing.T) {
	opts := jaeger.Options{Enabled: true}
	opts.Sampler.Type = "const"
	opts.Sampler.Param = 1
	opts.Reporter.Address = "192.168.99.100:6831"
	tracer, closer := jaeger.MustInit("test", opts)
	defer closer.Close()

	parent := newSpan(tracer, nil, "parent", time.Millisecond*100)
	defer parent.Finish()

	child := newSpan(tracer, parent, "child", time.Millisecond*200)
	defer child.Finish()
}

func newSpan(tracer opentracing.Tracer, parent opentracing.Span, name string, d time.Duration) (span opentracing.Span) {
	if parent == nil {
		span = tracer.StartSpan(name)
	} else {
		span = tracer.StartSpan(name, opentracing.ChildOf(parent.Context()))
	}
	time.Sleep(d)
	return
}

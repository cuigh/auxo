package jaeger_test

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/apm/trace/jaeger"
	"github.com/cuigh/auxo/test/assert"
)

func TestTrace(t *testing.T) {
	opts := jaeger.Options{Name: "test", Enabled: true}
	opts.Sampler.Type = "const"
	opts.Sampler.Param = 1
	opts.Reporter.Address = "192.168.99.100:6831"

	closer, err := jaeger.Init(opts)
	assert.NoError(t, err)
	defer closer.Close()

	parent := newSpan(nil, "parent", time.Millisecond*100)
	defer parent.Finish()

	child := newSpan(parent, "child", time.Millisecond*200)
	defer child.Finish()
}

func newSpan(parent trace.Span, name string, d time.Duration) (span trace.Span) {
	if parent == nil {
		span = trace.GetTracer().StartSpan(name)
	} else {
		span = trace.GetTracer().StartChild(parent, name)
	}
	time.Sleep(d)
	return
}

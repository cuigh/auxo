package trace_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/data/guid"
	"github.com/cuigh/auxo/test/assert"
)

type mockTracer struct {
}

func (mockTracer) Report(s *trace.Span) {
	stringer := func(b []byte) string {
		var id guid.ID
		copy(id[:], b)
		return id.String()
	}

	fmt.Printf(`span: {
	name: %s,
	id: %s,
	tid: %s,
	pid: %s,
	start: %s,
	end: %s,
	labels: %v,
}
`, s.Name, stringer(s.ID), stringer(s.TraceID), stringer(s.ParentID), s.Start, s.End, s.Labels)
}

func TestTrace(t *testing.T) {
	trace.SetTracer(mockTracer{})

	s := trace.NewRoot("/foo")
	time.Sleep(time.Second)
	s.Finish()

	trace.SetSampler(func() bool {
		return true
	})
	s = trace.NewRoot("/foo")
	time.Sleep(time.Second)
	s.Finish()

	s = s.NewChild("/bar")
	s.SetLabel("hello", "world")
	assert.Equal(t, "world", s.GetLabel("hello"))
	s.Finish()
}

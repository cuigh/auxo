package trace

import (
	"context"
	"math/rand"
	"time"

	"github.com/cuigh/auxo/data/guid"
)

var (
	ctxKey  = contextKey{}
	tracer  Tracer
	sampler Sampler = func() bool {
		return rand.Int31n(1000) < 10 // 1%
	}
	identifier Identifier = func() []byte {
		return guid.New().Slice()
	}
)

type contextKey struct{}

type Tracer interface {
	Report(s *Span)
}

type Sampler func() bool

type Identifier func() []byte

func SetTracer(t Tracer) {
	tracer = t
}

func SetSampler(s Sampler) {
	if s == nil {
		panic("sampler can't be nil")
	}
	sampler = s
}

func SetIdentifier(i Identifier) {
	if i == nil {
		panic("identifier can't be nil")
	}
	identifier = i
}

type Span struct {
	ID       []byte    `json:"id"`
	TraceID  []byte    `json:"tid"`
	ParentID []byte    `json:"pid"`
	Name     string    `json:"name"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Labels   map[string]string

	sample bool
}

func NewRoot(name string) *Span {
	if tracer == nil {
		return nil
	}

	return &Span{
		ID:      identifier(),
		TraceID: identifier(),
		Name:    name,
		Start:   time.Now(),
		sample:  sampler(),
	}
}

func NewChild(name string, tid, pid []byte) *Span {
	if tracer == nil {
		return nil
	}

	return &Span{
		ID:       identifier(),
		TraceID:  tid,
		ParentID: pid,
		Name:     name,
		Start:    time.Now(),
		sample:   true,
	}
}

func (s *Span) Finish() {
	if s == nil || !s.End.IsZero() {
		return
	}
	s.End = time.Now()

	if tracer != nil && s.sample {
		tracer.Report(s)
	}
}

func (s *Span) NewChild(name string) *Span {
	if s == nil {
		return nil
	}

	return &Span{
		ID:       identifier(),
		TraceID:  s.TraceID,
		ParentID: s.ID,
		Name:     name,
		Start:    time.Now(),
		sample:   s.sample,
	}
}

func (s *Span) SetLabel(k, v string) *Span {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	s.Labels[k] = v
	return s
}

func (s *Span) GetLabel(k string) string {
	if s.Labels == nil {
		return ""
	}
	return s.Labels[k]
}

func NewContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, ctxKey, span)
}

func FromContext(ctx context.Context) *Span {
	if v := ctx.Value(ctxKey); v != nil {
		return v.(*Span)
	}
	return nil
}

//func ToRequest(r *http.Request, span *Span) *http.Request {
//	return r.WithContext(NewContext(r.Context(), span))
//}
//
//func FromRequest(r *http.Request) *Span {
//	return FromContext(r.Context())
//}

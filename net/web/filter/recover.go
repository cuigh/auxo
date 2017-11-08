package filter

import (
	"runtime"
	"sync"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/net/web"
)

type Recover struct {
	// StackSize is max size of the stack to be logged.
	// Optional. Default value 4KB.
	StackSize int `json:"stack_size"`

	// StackAll enables logging stack traces of all other goroutines.
	// Optional. Default value false.
	StackAll bool `json:"stack_all"`

	// StackEnabled enables logging stack trace.
	// Optional. Default value is true.
	StackEnabled bool `json:"stack_enabled"`
}

// NewRecover returns a Recover instance.
func NewRecover() *Recover {
	return &Recover{
		StackEnabled: true,
	}
}

// Apply implements `web.Filter` interface.
func (r *Recover) Apply(next web.HandlerFunc) web.HandlerFunc {
	if r.StackSize == 0 {
		r.StackSize = 4 << 10 // 4 KB
	}

	pool := sync.Pool{
		New: func() interface{} {
			return make([]byte, r.StackSize)
		},
	}

	return func(ctx web.Context) error {
		defer func() {
			if e := recover(); e != nil {
				err := errors.Convert(e)
				if r.StackEnabled {
					stack := pool.Get().([]byte)
					length := runtime.Stack(stack, r.StackAll)
					ctx.Logger().Errorf("[%s] %s %s", "PANIC", err, stack[:length])
					pool.Put(stack)
				}
				ctx.Error(err)
			}
		}()

		return next(ctx)
	}
}

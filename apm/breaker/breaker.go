package breaker

import (
	"sync/atomic"
	"time"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
)

const PkgName = "auxo.apm.breaker"

var (
	// ErrMaxConcurrency occurs when too many of the same named command are executed at the same time.
	ErrMaxConcurrency = errors.New("breaker: max concurrency")
	// ErrCircuitOpen returns when an execution attempt "short circuits". This happens due to the circuit being measured as unhealthy.
	ErrCircuitOpen = errors.New("breaker: circuit open")
	// ErrTimeout occurs when the provided function takes too long to execute.
	ErrTimeout = errors.New("breaker: timeout")
)

// Valid CircuitState values.
const (
	Closed int32 = iota
	Open
	HalfOpen
)

type Action func() error
type Fallback func(error) error

// Options holds options of Breaker
type Options struct {
	// Window is how long to wait for state switched to half-open from open.
	Window time.Duration
	// Timeout is how long to wait for action to complete.
	Timeout time.Duration
	// Concurrency is how many actions can execute at the same time. If concurrency is zero, no limitation is set.
	Concurrency int32
}

type Breaker struct {
	name   string
	cond   Condition
	opts   Options
	logger log.Logger

	state  int32
	expiry time.Time
	active int32
	stat   statistics
}

func NewBreaker(name string, cond Condition, opts Options) *Breaker {
	if opts.Window <= 0 {
		opts.Window = time.Second * 5
	}
	//if opts.Timeout <= 0 {
	//	opts.Timeout = time.Second
	//}

	return &Breaker{
		name:   name,
		cond:   cond,
		opts:   opts,
		logger: log.Get(PkgName),
	}
}

func (b *Breaker) State() int32 {
	return b.state
}

func (b *Breaker) Summary() Summary {
	return b.stat.summary
}

func (b *Breaker) Try(action Action, fallback ...Fallback) error {
	switch b.state {
	case HalfOpen:
		return b.fail(ErrCircuitOpen)
	case Open:
		if time.Now().Before(b.expiry) {
			return b.fail(ErrCircuitOpen)
		}
		return b.test(action, fallback...)
	default: // Closed
		if b.opts.Concurrency > 0 {
			active := atomic.AddInt32(&b.active, 1)
			defer atomic.AddInt32(&b.active, -1)
			if active > b.opts.Concurrency {
				return b.fallback(ErrMaxConcurrency, fallback...)
			}
		}
		return b.call(action, fallback...)
	}
}

func (b *Breaker) call(action Action, fb ...Fallback) error {
	err := b.execute(action)
	if err == nil {
		b.stat.update(nil, nil, false)
		return nil
	}
	return b.fallback(err, fb...)
}

// let one action passing for testing
func (b *Breaker) test(action Action, fb ...Fallback) error {
	if !atomic.CompareAndSwapInt32(&b.state, Open, HalfOpen) {
		if b.state == Closed {
			return b.call(action, fb...)
		}
		return b.fail(ErrCircuitOpen)
	}

	b.logger.Info("breaker > half-open circuit: ", b.name)

	err := b.call(action, fb...)
	if err == nil {
		b.state = Closed
		b.stat.reset(false, true)
		b.logger.Info("breaker > close circuit: ", b.name)
	} else {
		b.state = Open
		b.expiry = time.Now().Add(b.opts.Window)
		b.logger.Warn("breaker > open circuit: ", b.name)
	}
	return err
}

func (b *Breaker) execute(action Action) error {
	if b.opts.Timeout > 0 {
		errs := make(chan error)
		go func() {
			errs <- action()
		}()

		timer := time.NewTimer(b.opts.Timeout)
		defer timer.Stop()

		select {
		case err := <-errs:
			return err
		case <-timer.C:
			return ErrTimeout
		}
	}
	return action()
}

func (b *Breaker) fail(err error) error {
	b.stat.update(ErrCircuitOpen, nil, false)
	return ErrCircuitOpen
}

// try call fallback function if any.
func (b *Breaker) fallback(err error, fb ...Fallback) error {
	var final error
	if len(fb) == 0 {
		final = err
		b.stat.update(err, nil, false)
	} else {
		final = fb[0](err)
		b.stat.update(final, err, true)
	}

	if final != nil && b.cond(b.stat.counter) && atomic.CompareAndSwapInt32(&b.state, Closed, Open) {
		b.expiry = time.Now().Add(b.opts.Window)
		b.stat.reset(false, true)
		b.logger.Warn("breaker > open circuit: ", b.name)
	}
	return final
}

//func (b *Breaker) Allow() bool {
//	switch b.state {
//	case HalfOpen:
//		return false
//	case Open:
//		if time.Now().Before(b.expiry) {
//			return false
//		}
//
//		if atomic.CompareAndSwapInt32(&b.state, Open, HalfOpen) {
//			return true
//		} else if b.state == Closed {
//			return true
//		}
//		return false
//	default: // Closed
//		return true
//	}
//}
//
//func (b *Breaker) Succeed() {
//	b.stat.update(nil, nil, false)
//	switch b.state {
//	case HalfOpen:
//		b.state = Closed
//	}
//}
//
//func (b *Breaker) Fail(err error) {
//	b.stat.update(err, nil, false)
//	switch b.state {
//	case HalfOpen:
//		b.state = Open
//	case Closed:
//		if b.cond(b.stat.counter) && atomic.CompareAndSwapInt32(&b.state, Closed, Open) {
//			b.stat.reset(false, true)
//		}
//	}
//}

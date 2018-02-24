package run

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/util/debug"
)

const PkgPath = "auxo.util.run"

var (
	ErrTimeout      = errors.New("timeout")
	ErrPoolShutdown = errors.New("pool is shut down")
	ErrPoolFull     = errors.New("pool reach max queue size")
)

type Recovery func(e interface{})

type Canceler interface {
	Cancel()
}

func handlePanic(r Recovery, e interface{}) {
	if r == nil {
		log.Get(PkgPath).Errorf("PANIC: %v, stack: %s", e, debug.StackSkip(1))
	} else {
		r(e)
	}
}

// Safe call fn with recover.
func Safe(fn func(), r Recovery) {
	defer func() {
		if e := recover(); e != nil {
			handlePanic(r, e)
		}
	}()
	fn()
}

// Count call fn with recover, and process WaitGroup automatically.
func Count(g *sync.WaitGroup, fn func(), r Recovery) {
	g.Add(1)
	defer func() {
		g.Done()
		if e := recover(); e != nil {
			handlePanic(r, e)
		}
	}()
	fn()
}

// Pipeline calls all functions in order. If a function returns an error, Pipeline will return this error immediately.
func Pipeline(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// Schedule call fn with recover continuously. It returns a Canceler that can
// be used to cancel the call using its Cancel method.
func Schedule(d time.Duration, fn func(), r Recovery) Canceler {
	s := &scheduler{
		d:     d,
		fn:    fn,
		r:     r,
		state: 1,
	}
	s.timer = time.AfterFunc(d, s.schedule)
	return s
}

type scheduler struct {
	d     time.Duration
	fn    func()
	r     Recovery
	timer *time.Timer
	state int32
}

func (s *scheduler) schedule() {
	if s.state == 1 {
		Safe(s.fn, s.r)
		s.timer = time.AfterFunc(s.d, s.schedule)
	}
}

func (s *scheduler) Cancel() {
	if atomic.CompareAndSwapInt32(&s.state, 1, 0) && s.timer != nil {
		s.timer.Stop()
	}
}

// Pool is a simple goroutine pool.
type Pool struct {
	Min, Max int32
	Backlog  int
	Idle     time.Duration

	state    int32
	current  int32 // running goroutines
	busy     int32 // busy goroutines
	jobs     chan func()
	stopper  data.Chan // stop all goroutine
	closer   data.Chan // close idle goroutine
	canceler Canceler  // control canceler
}

// Start start the pool
func (p *Pool) Start() {
	if atomic.CompareAndSwapInt32(&p.state, 0, 1) {
		if p.Min <= 0 {
			p.Min = 1
		}
		if p.Max <= 0 {
			p.Max = 100000
		} else if p.Max < p.Min {
			p.Max = p.Min
		}
		if p.Backlog <= 0 {
			p.Backlog = 10000
		}
		if p.Idle <= 0 {
			p.Idle = time.Minute
		}
		p.jobs = make(chan func(), p.Backlog)
		p.stopper = make(data.Chan)
		p.closer = make(data.Chan)
		for i := int32(0); i < p.Min; i++ {
			go p.long()
		}
		p.control()
	}
}

// Stop shuts down the pool. Stop won't interrupt the running jobs.
func (p *Pool) Stop() {
	if atomic.CompareAndSwapInt32(&p.state, 1, 0) {
		close(p.stopper)
		if p.canceler != nil {
			p.canceler.Cancel()
		}
	}
}

// Wait waits the running jobs to finish.
func (p *Pool) Wait(timeout time.Duration) error {
	const waitPollInterval = 500 * time.Millisecond

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	ticker := time.NewTicker(waitPollInterval)
	defer ticker.Stop()

	for {
		if p.busy == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// Put submits the job to pool.
func (p *Pool) Put(job func()) error {
	if atomic.LoadInt32(&p.state) != 1 {
		return ErrPoolShutdown
	}

	select {
	case p.jobs <- job:
		if p.busy >= p.current && p.current < p.Max {
			go p.short()
		}
		return nil
	default:
		return ErrPoolFull
	}
}

func (p *Pool) long() {
	atomic.AddInt32(&p.current, 1)
	defer atomic.AddInt32(&p.current, -1)

	for {
		select {
		case job := <-p.jobs:
			p.call(job)
		case <-p.stopper:
			return
		}
	}
}

func (p *Pool) short() {
	atomic.AddInt32(&p.current, 1)
	defer atomic.AddInt32(&p.current, -1)

	t := time.NewTimer(p.Idle)
	defer t.Stop()

	for {
		select {
		case job := <-p.jobs:
			p.call(job)
			t.Reset(p.Idle)
		case <-p.stopper:
			return
		case <-p.closer:
			//log.Get(PkgPath).Debugf("run > Pool: Close goroutine by scheduler[%d/%d]", p.busy, p.current)
			return
		case <-t.C:
			//log.Get(PkgPath).Debugf("run > Pool: Close goroutine for idle timeout[%d/%d]", p.busy, p.current)
			return
		}
	}
}

// Adjust the number of goroutines to conserve resources.
func (p *Pool) control() {
	p.canceler = Schedule(5*time.Second, func() {
		if p.busy < p.current/2 && p.current > p.Min {
			p.closer.TrySend()
		}
	}, nil)
}

func (p *Pool) call(fn func()) {
	atomic.AddInt32(&p.busy, 1)
	defer atomic.AddInt32(&p.busy, -1)

	fn()
}

package run

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrTimeout      = errors.New("timeout")
	ErrPoolShutdown = errors.New("pool is shut down")
	ErrPoolFull     = errors.New("pool reach max queue size")
)

type Recovery func(e interface{})

// Safe call fn with recover.
func Safe(fn func(), r Recovery) {
	defer func() {
		if e := recover(); e != nil {
			r(e)
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
			r(e)
		}
	}()
	fn()
}

// Pool is a simple goroutine pool.
type Pool struct {
	Min, Max int32
	Backlog  int
	Idle     time.Duration

	state   int32
	current int32 // running goroutines
	jobs    chan func()
	stopper chan struct{}
	wg      sync.WaitGroup
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
			p.Idle = 3 * time.Minute
		}
		p.jobs = make(chan func(), p.Backlog)
		p.stopper = make(chan struct{})
		for i := int32(0); i < p.Min; i++ {
			go p.long()
		}
	}
}

// Stop shuts down the pool. Stop won't interrupt the running jobs.
func (p *Pool) Stop() {
	if atomic.CompareAndSwapInt32(&p.state, 1, 0) {
		close(p.stopper)
	}
}

// Wait waits the running jobs to finish.
func (p *Pool) Wait(timeout time.Duration) error {
	ch := make(chan struct{})
	go func() {
		p.wg.Wait()
		ch <- struct{}{}
	}()

	select {
	case <-time.After(timeout):
		return ErrTimeout
	case <-ch:
		return nil
	}
}

// Put submits the job to pool.
func (p *Pool) Put(job func()) error {
	if atomic.LoadInt32(&p.state) != 1 {
		return ErrPoolShutdown
	}

	select {
	case p.jobs <- job:
		if len(p.jobs) > 1 && p.current < p.Max {
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
		case <-t.C:
			return
		}
	}
}

func (p *Pool) call(fn func()) {
	p.wg.Add(1)
	defer p.wg.Done()

	fn()
}

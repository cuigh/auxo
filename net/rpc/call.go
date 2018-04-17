package rpc

import (
	ct "context"
	"sync"
	"sync/atomic"

	"github.com/cuigh/auxo/data"
)

// CFilter is client interceptor.
type CFilter func(CHandler) CHandler

type CHandler func(*Call) error

type Call struct {
	n     *Node
	req   *Request
	ctx   ct.Context
	reply interface{}
	err   chan error
	//done  data.Chan
	released uint32
	next     *Call
}

func (c *Call) Server() string {
	return c.n.c.opts.Name
}

func (c *Call) Request() *Request {
	return c.req
}

func (c *Call) Context() ct.Context {
	return c.ctx
}

func (c *Call) SetContext(ctx ct.Context) {
	c.ctx = ctx
}

func (c *Call) Reply() interface{} {
	return c.reply
}

func (c *Call) Err() error {
	return nil
}

// Wait implements interface of AsyncError.
func (c *Call) Wait() (err error) {
	select {
	case err = <-c.err:
	case <-c.ctx.Done():
		if err = c.ctx.Err(); err == ct.Canceled {
			err = NewError(StatusCanceled, err.Error())
		} else {
			err = NewError(StatusDeadlineExceeded, err.Error())
		}
	}
	//c.n.calls.Release(c)
	return
}

func (c *Call) finish(err error) {
	c.err <- err
	//c.done.TrySend()
}

func (c *Call) reset(ctx ct.Context, service, method string, args []interface{}, reply interface{}) {
	//c.Head.Type = 0
	//c.req.Head.ID = id
	c.released = 0
	c.req.Head.Service = service
	c.req.Head.Method = method
	if c.req.Head.Labels != nil {
		c.req.Head.Labels = c.req.Head.Labels[:0]
	}
	c.req.Args = args
	c.ctx = ctx
	c.reply = reply
	c.err = make(chan error, 1)
	//c.err = nil
	//c.done = make(data.Chan, 1)
}

type callPool struct {
	build func() *Call

	sync.RWMutex
	free    *Call
	pending map[uint64]*Call
}

func newCallPool(n *Node) *callPool {
	cp := &callPool{
		build: func() *Call {
			return &Call{n: n, req: &Request{
				Head: RequestHead{
					Labels: make(data.Options, 0),
				},
			}}
		},
		pending: make(map[uint64]*Call),
	}
	return cp
}

// acquire a *Call from pool and add it to pending.
func (cp *callPool) Acquire(id uint64) *Call {
	cp.Lock()
	c := cp.free
	if c == nil {
		c = cp.build()
	} else {
		cp.free = c.next
	}
	c.req.Head.ID = id
	cp.pending[id] = c
	cp.Unlock()
	return c
}

// release a *Call to pool and remove it from pending.
func (cp *callPool) Release(c *Call) {
	if c.released == 0 && atomic.AddUint32(&c.released, 1) == 1 {
		cp.Lock()
		c.next = cp.free
		cp.free = c
		delete(cp.pending, c.req.Head.ID)
		cp.Unlock()
	}
}

// find pending *Call by id.
func (cp *callPool) Find(id uint64) (c *Call) {
	cp.RLock()
	c = cp.pending[id]
	cp.RUnlock()
	return
}

// clear all pending *Calls.
func (cp *callPool) Clear(fn func(c *Call)) {
	cp.Lock()
	for k, c := range cp.pending {
		fn(c)
		delete(cp.pending, k)
	}
	cp.Unlock()
}

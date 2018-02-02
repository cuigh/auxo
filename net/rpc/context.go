package rpc

import (
	ct "context"
	"sync"

	"github.com/cuigh/auxo/security"
)

var ctxKey = contextKey{}

type contextKey struct{}

type Context interface {
	Request() *Request
	Action() Action
	Context() ct.Context
	SetContext(ctx ct.Context)
	// User returns info of current visitor.
	User() security.User
	// User set user info of current visitor. Generally used by authentication filter.
	SetUser(user security.User)
}

func FromContext(ctx ct.Context) Context {
	if v := ctx.Value(ctxKey); v != nil {
		return v.(Context)
	}
	return nil
}

type context struct {
	svr    *Server
	ch     *Channel
	codec  ServerCodec
	req    *Request
	resp   *Response
	action Action
	ctx    ct.Context
}

func (c *context) User() security.User {
	return c.ch.u
}

func (c *context) SetUser(user security.User) {
	c.ch.u = user
}

//func (*context) Deadline() (deadline time.Time, ok bool) {
//	panic("implement me")
//}
//
//func (*context) Done() <-chan struct{} {
//	panic("implement me")
//}
//
//func (*context) Err() error {
//	panic("implement me")
//}
//
//func (c *context) Value(key interface{}) interface{} {
//  c.Request.Labels.Get(key)
//	panic("implement me")
//}

func (c *context) Context() ct.Context {
	if c.ctx == nil {
		return ct.Background()
	}
	return c.ctx
}

func (c *context) SetContext(ctx ct.Context) {
	c.ctx = ctx
}

func (c *context) Request() *Request {
	return c.req
}

func (c *context) Action() Action {
	return c.action
}

//func (c *context) respond(r interface{}, err error) {
//	c.resp.Head.ID = c.req.Head.ID
//	c.resp.Head.Type = 0
//	if err == nil {
//		c.resp.Result.Value = r
//	} else {
//		c.resp.Result.Error = wrapError(err)
//	}
//
//	err = c.codec.Encode(c.resp)
//	if err != nil {
//		s.logger.Error("encode response failed: ", err)
//	} else {
//		c.ch.Flush()
//	}
//}

func (c *context) Reset(ch *Channel, sc ServerCodec) {
	c.ch = ch
	c.codec = sc
	c.action = nil
	c.ctx = ct.Background()
	c.req.Head.ID = 0
	c.req.Head.Service = ""
	c.req.Head.Method = ""
	if c.req.Head.Labels != nil {
		c.req.Head.Labels = c.req.Head.Labels[:0]
	}
	c.req.Args = nil
	c.resp.Result.Value = nil
	c.resp.Result.Error = nil
}

type contextPool struct {
	sync.Pool
}

func newContextPool(s *Server) *contextPool {
	p := &contextPool{}
	p.New = func() interface{} {
		return &context{
			svr:  s,
			ctx:  ct.Background(),
			req:  &Request{},
			resp: &Response{},
		}
	}
	return p
}

func (p *contextPool) Get(ch *Channel, sc ServerCodec) (c *context) {
	c = p.Pool.Get().(*context)
	c.Reset(ch, sc)
	return
}

func (p *contextPool) Put(ctx *context) {
	p.Pool.Put(ctx)
}

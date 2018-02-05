package proto

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/cuigh/auxo/byte/buffer"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/net/rpc"
	"github.com/cuigh/auxo/util/cast"
	"github.com/gogo/protobuf/proto"
)

const (
	PkgName               = "auxo.net.rpc.codec.proto"
	DefaultMaxMessageSize = 2 << 20
)

//var header = []byte("proto")

func (m *Request) reset() {
	if m.Args != nil {
		m.Args = m.Args[:0]
	}
	if m.Labels != nil {
		m.Labels = m.Labels[:0]
	}
}

func (m *Response) reset() {
	m.Result = nil
	m.Error = nil
}

type ClientCodec struct {
	rpc.Stream
	req        *Request
	resp       *Response
	sendBuf    *proto.Buffer
	receiveBuf *proto.Buffer
	bufPool    *buffer.GroupPool
	maxMsgSize int
	lock       sync.Mutex // protect for writing
}

func (c *ClientCodec) Encode(req *rpc.Request) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.req.ID = req.Head.ID
	c.req.Service = req.Head.Service
	c.req.Method = req.Head.Method
	if l := len(req.Args); l > 0 {
		c.req.Args = make([][]byte, l)
		for i, arg := range req.Args {
			d, err := proto.Marshal(arg.(proto.Message))
			if err != nil {
				return err
			}
			c.req.Args[i] = d
		}
	}
	if l := len(req.Head.Labels); l > 0 {
		c.req.Labels = make([]*Label, l)
		for i, label := range req.Head.Labels {
			c.req.Labels[i] = (*Label)(&label)
		}
	}
	err = c.sendBuf.Marshal(c.req)
	if err == nil {
		err = c.write(c.sendBuf.Bytes())
	}
	c.sendBuf.Reset()
	return err
}

func (c *ClientCodec) write(data []byte) error {
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	c.Write(size)
	c.Write(data)
	return c.Flush()
}

func (c *ClientCodec) DecodeHead(head *rpc.ResponseHead) error {
	// [4 bit length] + [data]
	var (
		err    error
		length int
	)

	size := make([]byte, 4)
	_, err = io.ReadFull(c, size)
	if err != nil {
		return err
	}

	length = int(binary.LittleEndian.Uint32(size))
	if length > c.maxMsgSize {
		return errors.Format("message too big: %d", length)
	}

	// read body
	b := c.bufPool.Get(length)
	//defer c.bufPool.Put(b)
	buf := b[:length]
	_, err = io.ReadFull(c, buf)
	if err != nil {
		c.bufPool.Put(b)
		return err
	}

	// unmarshal message
	c.receiveBuf.SetBuf(buf)
	c.resp.reset()
	err = c.receiveBuf.Unmarshal(c.resp)
	c.bufPool.Put(b)
	if err != nil {
		return err
	}

	head.ID = c.resp.ID
	return nil
}

func (c *ClientCodec) DecodeResult(result *rpc.Result) (err error) {
	if c.resp.Error == nil {
		result.Error = nil
		if c.resp.Result != nil {
			err = proto.Unmarshal(c.resp.Result, result.Value.(proto.Message))
		}
	} else {
		result.Error = (*errors.CodedError)(c.resp.Error)
	}
	return
}

func (*ClientCodec) DiscardResult() error {
	return nil
}

type ServerCodec struct {
	rpc.Stream
	req        *Request
	resp       *Response
	sendBuf    *proto.Buffer
	receiveBuf *proto.Buffer
	bufPool    *buffer.GroupPool
	maxMsgSize int
	lock       sync.Mutex // protect for writing
}

func (c *ServerCodec) Encode(resp *rpc.Response) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.resp.ID = resp.Head.ID
	if resp.Result.Error == nil {
		d, err := proto.Marshal(resp.Result.Value.(proto.Message))
		if err != nil {
			return err
		}
		c.resp.Result = d
		c.resp.Error = nil
	} else {
		c.resp.Result = nil
		c.resp.Error = (*Error)(resp.Result.Error)
	}

	err = c.sendBuf.Marshal(c.resp)
	if err == nil {
		err = c.write(c.sendBuf.Bytes())
	}
	c.sendBuf.Reset()
	return
}

func (c *ServerCodec) write(data []byte) error {
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	c.Write(size)
	c.Write(data)
	return c.Flush()
}

func (c *ServerCodec) DecodeHead(head *rpc.RequestHead) error {
	// [4 bit length] + [data]
	var (
		err    error
		length int
	)

	size := make([]byte, 4)
	_, err = io.ReadFull(c, size)
	if err != nil {
		return err
	}

	length = int(binary.LittleEndian.Uint32(size))
	if length > c.maxMsgSize {
		return errors.Format("message too big: %d", length)
	}

	// read body
	b := c.bufPool.Get(length)
	//defer c.bufPool.Put(b)
	buf := b[:length]
	_, err = io.ReadFull(c, buf)
	if err != nil {
		c.bufPool.Put(b)
		return err
	}

	// unmarshal message
	c.receiveBuf.SetBuf(buf)
	c.req.reset()
	err = c.receiveBuf.Unmarshal(c.req)
	c.bufPool.Put(b)
	if err != nil {
		return err
	}

	head.ID = c.req.ID
	head.Service = c.req.Service
	head.Method = c.req.Method
	for _, l := range c.req.Labels {
		head.Labels = append(head.Labels, (data.Option)(*l))
	}
	return nil
}

func (c *ServerCodec) DecodeArgs(args []interface{}) error {
	for i, arg := range c.req.Args {
		err := proto.Unmarshal(arg, args[i].(proto.Message))
		if err != nil {
			return err
		}
	}
	return nil
}

func (*ServerCodec) DiscardArgs() error {
	return nil
}

// Matcher is an protocol buffer codec Matcher.
//func Matcher(p rpc.ReadPeeker) bool {
//	b, err := p.Peek(len(bytesRequestHeader))
//	return err == nil && bytes.Equal(b, bytesRequestHeader)
//}

type Builder struct {
}

func (b Builder) NewClient(s rpc.Stream, opts data.Map) rpc.ClientCodec {
	maxSize := b.maxMsgSize(opts)
	return &ClientCodec{
		Stream:     s,
		req:        &Request{},
		resp:       &Response{},
		sendBuf:    proto.NewBuffer(make([]byte, 0, 4<<10)),
		receiveBuf: proto.NewBuffer(make([]byte, 0, 4<<10)),
		maxMsgSize: maxSize,
		bufPool:    buffer.NewGroupPool(4<<10, maxSize, 2), // 4KB -> 2MB
	}
}

func (b Builder) NewServer(s rpc.Stream, opts data.Map) rpc.ServerCodec {
	maxSize := b.maxMsgSize(opts)
	return &ServerCodec{
		Stream:     s,
		req:        &Request{},
		resp:       &Response{},
		sendBuf:    proto.NewBuffer(make([]byte, 0, 4<<10)),
		receiveBuf: proto.NewBuffer(make([]byte, 0, 4<<10)),
		maxMsgSize: maxSize,
		bufPool:    buffer.NewGroupPool(4<<10, maxSize, 2), // 4KB -> 2MB
	}
}

func (b Builder) maxMsgSize(opts data.Map) (size int) {
	if opts != nil {
		size = cast.ToInt(opts.Get("max_msg_size"))
	}
	if size <= 0 {
		size = DefaultMaxMessageSize
	}
	return size
}

func init() {
	rpc.RegisterCodec("proto", Builder{})
}

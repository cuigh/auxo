package jsoniter

import (
	"sync"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/net/rpc"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type request struct {
	Head rpc.RequestHead       `json:"head"`
	Args []jsoniter.RawMessage `json:"args"`
}

func (r *request) reset() {
	r.Head.ID = 0
	r.Head.Service = ""
	r.Head.Method = ""
	if r.Head.Labels != nil {
		r.Head.Labels = r.Head.Labels[:0]
	}
	if r.Args != nil {
		r.Args = r.Args[:0]
	}
}

type response struct {
	Head   rpc.ResponseHead `json:"head"`
	Result struct {
		Value jsoniter.RawMessage `json:"value"`
		Error *errors.CodedError  `json:"error"`
	} `json:"result"`
}

func (r *response) reset() {
	r.Head.ID = 0
	r.Result.Value = nil
	r.Result.Error = nil
}

type ClientCodec struct {
	rpc.Stream
	enc  *jsoniter.Encoder
	dec  *jsoniter.Decoder
	resp *response
	lock sync.Mutex // protect for writing
}

func (c *ClientCodec) Encode(req *rpc.Request) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	err = c.enc.Encode(req)
	if err == nil {
		err = c.Flush()
	}
	return
}

func (c *ClientCodec) DecodeHead(head *rpc.ResponseHead) error {
	c.resp.reset()
	err := c.dec.Decode(c.resp)
	if err != nil {
		return err
	}
	*head = c.resp.Head
	return nil
}

func (c *ClientCodec) DecodeResult(result *rpc.Result) error {
	result.Error = c.resp.Result.Error
	if result.Value == nil {
		return nil
	}
	return json.Unmarshal(c.resp.Result.Value, result.Value)
}

func (*ClientCodec) DiscardResult() error {
	return nil
}

type ServerCodec struct {
	rpc.Stream
	enc  *jsoniter.Encoder
	dec  *jsoniter.Decoder
	req  *request
	lock sync.Mutex // protect for writing
}

func (c *ServerCodec) Encode(resp *rpc.Response) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	err = c.enc.Encode(resp)
	if err == nil {
		err = c.Flush()
	}
	return
}

func (c *ServerCodec) DecodeHead(head *rpc.RequestHead) error {
	c.req.reset()
	err := c.dec.Decode(c.req)
	if err != nil {
		return err
	}
	*head = c.req.Head
	return nil
}

func (c *ServerCodec) DecodeArgs(args []interface{}) error {
	for i, arg := range c.req.Args {
		err := json.Unmarshal(arg, args[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (*ServerCodec) DiscardArgs() error {
	return nil
}

// Matcher is a JSON Matcher.
func Matcher(p rpc.ReadPeeker) bool {
	b, err := p.Peek(1)
	return err == nil && b[0] == '{'
}

type Builder struct {
}

func (Builder) NewClient(s rpc.Stream, opts data.Map) rpc.ClientCodec {
	return &ClientCodec{
		Stream: s,
		enc:    json.NewEncoder(s),
		dec:    json.NewDecoder(s),
		resp:   &response{},
	}
}

func (Builder) NewServer(s rpc.Stream, opts data.Map) rpc.ServerCodec {
	return &ServerCodec{
		Stream: s,
		enc:    json.NewEncoder(s),
		dec:    json.NewDecoder(s),
		req:    &request{},
	}
}

func init() {
	rpc.RegisterCodec("json", Builder{})
}

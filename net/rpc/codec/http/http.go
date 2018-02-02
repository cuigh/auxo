package http

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/net/rpc"
	"github.com/cuigh/auxo/util/cast"
)

const (
	headerID = "AUXO-RPC-ID"
)

type ClientCodec struct {
	rpc.Stream
	path  string
	enc   *jsonEncoder
	req   *http.Request
	dummy *http.Request
	resp  *http.Response
	lock  sync.Mutex // protect for writing
}

// Encode send request to RPC server. For example
//
// 	POST /Test.Hello HTTP/1.1
// 	AUXO-RPC-ID: 5a3a470453d38640fa000001
//	Content-Length: 8
//
//	["auxo"]
func (c *ClientCodec) Encode(req *rpc.Request) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.enc.Reset()
	err = c.enc.Encode(req.Args)
	if err != nil {
		return err
	}

	c.req.URL.Path = texts.Concat(c.path, req.Head.Service, ".", req.Head.Method)
	c.req.Body = ioutil.NopCloser(c.enc)
	c.req.ContentLength = int64(c.enc.Len())
	c.req.Header.Set(headerID, strconv.FormatUint(req.Head.ID, 10))
	err = c.req.Write(c)
	if err == nil {
		err = c.Flush()
	}
	return err
}

func (c *ClientCodec) DecodeHead(head *rpc.ResponseHead) (err error) {
	c.resp, err = http.ReadResponse(c.Reader(), c.dummy)
	if err != nil {
		return
	}

	head.ID, err = strconv.ParseUint(c.resp.Header.Get(headerID), 10, 64)
	return
}

func (c *ClientCodec) DecodeResult(result *rpc.Result) error {
	defer c.resp.Body.Close()
	return json.NewDecoder(c.resp.Body).Decode(result)
}

func (c *ClientCodec) DiscardResult() (err error) {
	defer c.resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, c.resp.Body)
	return
}

type ServerCodec struct {
	rpc.Stream
	path string
	enc  *jsonEncoder
	resp *http.Response
	req  *http.Request
	lock sync.Mutex // protect for writing
}

// Encode send response to RPC client. For example
//
// HTTP/1.1 200 OK
// Content-Length: 24
// Auxo-Rpc-Id: 5a3a470453d38640fa000001
// Connection: Keep-Alive
// Content-Type: application/json; charset=UTF-8
// Date: Thu, 21 Dec 2017 11:18:52 CST
// Server: auxo-rpc
//
// {"value":"Hello, auxo"}
func (c *ServerCodec) Encode(resp *rpc.Response) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.enc.Reset()
	err = c.enc.Encode(resp.Result)
	if err != nil {
		return err
	}

	c.resp.Body = ioutil.NopCloser(c.enc)
	c.resp.ContentLength = int64(c.enc.Len())
	c.resp.Header.Set(headerID, strconv.FormatUint(resp.Head.ID, 10))
	err = c.resp.Write(c)
	if err == nil {
		err = c.Flush()
	}
	return
}

func (c *ServerCodec) DecodeHead(head *rpc.RequestHead) (err error) {
	c.req, err = http.ReadRequest(c.Reader())
	if err != nil {
		return
	}

	head.ID, err = strconv.ParseUint(c.req.Header.Get(headerID), 10, 64)
	if err != nil {
		return
	}

	slice := strings.SplitN(c.req.RequestURI[len(c.path):], ".", 2)
	head.Service = slice[0]
	head.Method = slice[1]
	return nil
}

func (c *ServerCodec) DecodeArgs(args []interface{}) error {
	slice := make([]json.RawMessage, 0)
	err := json.NewDecoder(c.req.Body).Decode(&slice)
	if err != nil {
		return err
	}

	for i, arg := range args {
		err := json.Unmarshal(slice[i], arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ServerCodec) DiscardArgs() error {
	_, err := io.Copy(ioutil.Discard, c.req.Body)
	return err
}

// Matcher is an HTTP Matcher.
func Matcher(p rpc.ReadPeeker) bool {
	b, err := p.Peek(5)
	if err != nil {
		return false
	}
	return bytes.Equal(b, []byte("POST "))
}

type Builder struct {
}

func (Builder) NewClient(s rpc.Stream, opts data.Map) rpc.ClientCodec {
	var p string
	if opts != nil {
		p = cast.ToString(opts.Get("path"))
	}
	if p == "" {
		p = "/"
	}
	return &ClientCodec{
		Stream: s,
		path:   p,
		enc:    newJSONEncoder(),
		req: &http.Request{
			Method:     http.MethodPost,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			URL:        &url.URL{},
			Header:     make(http.Header),
		},
		dummy: &http.Request{Method: http.MethodPost},
	}
}

func (Builder) NewServer(s rpc.Stream, opts data.Map) rpc.ServerCodec {
	var p string
	if opts != nil {
		p = cast.ToString(opts.Get("path"))
	}
	if p == "" {
		p = "/"
	}
	return &ServerCodec{
		Stream: s,
		path:   p,
		enc:    newJSONEncoder(),
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				// todo: set header according to client?
				"Connection":   []string{"Keep-Alive"},
				"Server":       []string{"auxo-rpc"},
				"Date":         []string{time.Now().Format(time.RFC1123)},
				"Content-Type": []string{"application/json; charset=UTF-8"},
			},
		},
	}
}

type jsonEncoder struct {
	*bytes.Buffer
	*json.Encoder
}

func newJSONEncoder() *jsonEncoder {
	e := &jsonEncoder{
		Buffer: &bytes.Buffer{},
	}
	e.Encoder = json.NewEncoder(e.Buffer)
	return e
}

func init() {
	rpc.RegisterCodec("http", Builder{})
}

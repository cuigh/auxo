package rpc

import (
	"bufio"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
)

var (
	codecs = map[string]CodecBuilder{}
)

func RegisterCodec(name string, cb CodecBuilder) {
	codecs[name] = cb
}

type Stream interface {
	// read
	Reader() *bufio.Reader
	Peek(n int) ([]byte, error)
	Read(p []byte) (n int, err error)
	ReadByte() (byte, error)
	ReadBytes(delim byte) ([]byte, error)
	ReadString(delim byte) (string, error)
	//func (b *Reader) Buffered() int
	//func (b *Reader) Discard(n int) (discarded int, err error)
	//func (b *Reader) Peek(n int) ([]byte, error)
	//func (b *Reader) Read(p []byte) (n int, err error)
	//func (b *Reader) ReadByte() (byte, error)
	//func (b *Reader) ReadBytes(delim byte) ([]byte, error)
	//func (b *Reader) ReadLine() (line []byte, isPrefix bool, err error)
	//func (b *Reader) ReadRune() (r rune, size int, err error)
	//func (b *Reader) ReadSlice(delim byte) (line []byte, err error)
	//func (b *Reader) ReadString(delim byte) (string, error)
	//func (b *Reader) Reset(r io.Reader)
	//func (b *Reader) UnreadByte() error
	//func (b *Reader) UnreadRune() error
	//func (b *Reader) WriteTo(w io.Writer) (n int64, err error)

	// write
	Writer() *bufio.Writer
	Write(p []byte) (n int, err error)
	WriteByte(c byte) error
	WriteString(s string) (int, error)
	Flush() error
	//func (b *Writer) Available() int
	//func (b *Writer) Buffered() int
	//func (b *Writer) Flush() error
	//func (b *Writer) ReadFrom(r io.Reader) (n int64, err error)
	//func (b *Writer) Reset(w io.Writer)
	//func (b *Writer) Write(p []byte) (nn int, err error)
	//func (b *Writer) WriteByte(c byte) error
	//func (b *Writer) WriteRune(r rune) (size int, err error)
	//func (b *Writer) WriteString(s string) (int, error)
}

type RequestHead struct {
	//Type    byte         `json:"type"` // 0-rpc, 1-heartbeat
	ID      uint64       `json:"id"` // len(ID) == 0 for a heartbeat response
	Service string       `json:"service,omitempty"`
	Method  string       `json:"method,omitempty"`
	Labels  data.Options `json:"labels,omitempty"`
	//TraceID []byte
}

type Request struct {
	Head RequestHead   `json:"head"`
	Args []interface{} `json:"args,omitempty"`
}

type ResponseHead struct {
	//Type byte   `json:"type"` // 0-rpc, 1-heartbeat
	ID uint64 `json:"id"` // len(ID) == 0 for a heartbeat request
}

type Result struct {
	Value interface{}        `json:"value,omitempty"`
	Error *errors.CodedError `json:"error,omitempty"`
}

type Response struct {
	Head   ResponseHead `json:"head"`
	Result Result       `json:"result"`
}

type ClientCodec interface {
	// Encode send request to server, must be concurrent safe.
	Encode(req *Request) error
	DecodeHead(head *ResponseHead) error
	DecodeResult(result *Result) error
	DiscardResult() error
}

type ServerCodec interface {
	// Encode send response to client, must be concurrent safe.
	Encode(resp *Response) error
	DecodeHead(head *RequestHead) error
	DecodeArgs(args []interface{}) error
	DiscardArgs() error
}

type CodecBuilder interface {
	NewClient(s Stream, opts data.Map) ClientCodec
	NewServer(s Stream, opts data.Map) ServerCodec
}

type ReadPeeker interface {
	Read(p []byte) (n int, err error)
	Peek(n int) ([]byte, error)
}

// Matcher matches a connection based on it's content.
type Matcher func(p ReadPeeker) bool

// Any is an always matched Matcher.
func Any(_ ReadPeeker) bool {
	return true
}

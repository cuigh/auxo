package data

import (
	"errors"
	"strings"
)

var (
	// Empty is a empty struct instance.
	Empty             = struct{}{}
	Nil         Value = nilValue{}
	ErrNilValue       = errors.New("nil value")
)

type Value interface {
	IsNil() bool
	Scan(i interface{}) error
	Bytes() ([]byte, error)
	Bool() (bool, error)
	String() (string, error)
	Int() (int, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)
	Uint() (uint, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	//Time() (time.Time, error)
	//Duration() (time.Duration, error)
}

type nilValue struct {
}

func (nilValue) IsNil() bool {
	return true
}

func (nilValue) Scan(i interface{}) error {
	return ErrNilValue
}

func (nilValue) Bytes() ([]byte, error) {
	return nil, ErrNilValue
}

func (nilValue) Bool() (bool, error) {
	return false, ErrNilValue
}

func (nilValue) String() (string, error) {
	return "", ErrNilValue
}

func (nilValue) Int() (int, error) {
	return 0, ErrNilValue
}

func (nilValue) Int8() (int8, error) {
	return 0, ErrNilValue
}

func (nilValue) Int16() (int16, error) {
	return 0, ErrNilValue
}

func (nilValue) Int32() (int32, error) {
	return 0, ErrNilValue
}

func (nilValue) Int64() (int64, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint() (uint, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint8() (uint8, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint16() (uint16, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint32() (uint32, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint64() (uint64, error) {
	return 0, ErrNilValue
}

func (nilValue) Float32() (float32, error) {
	return 0, ErrNilValue
}

func (nilValue) Float64() (float64, error) {
	return 0, ErrNilValue
}

type Option struct {
	Name  string `json:"name" xml:"name,attr"`
	Value string `json:"value" xml:"value,attr"`
}

func ParseOption(s, sep string) Option {
	pair := strings.SplitN(s, sep, 2)
	return Option{
		Name:  pair[0],
		Value: pair[1],
	}
}

type Options []Option

func ParseOptions(s, sep1, sep2 string) Options {
	parts := strings.Split(s, sep1)
	opts := make(Options, len(parts))
	for i, part := range parts {
		opts[i] = ParseOption(part, sep2)
	}
	return opts
}

func (opts Options) Get(name string) string {
	for _, opt := range opts {
		if opt.Name == name {
			return opt.Value
		}
	}
	return ""
}

// Chan is a simple notification channel
type Chan chan struct{}

// ReadOnly converts c to receive-only channel
func (c Chan) ReadOnly() ReadChan {
	return (chan struct{})(c)
}

// WriteOnly converts c to send-only channel
func (c Chan) WriteOnly() WriteChan {
	return (chan struct{})(c)
}

// Send send a notification to channel. It blocks if channel is full.
func (c Chan) Send() {
	c <- Empty
}

// Send send a notification to channel. It returns false if channel is full.
func (c Chan) TrySend() bool {
	select {
	case c <- Empty:
		return true
	default:
		return false
	}
}

// Receive receive a notification from channel. It blocks if channel is empty.
func (c Chan) Receive() {
	<-c
}

// Receive receive a notification from channel. It returns false if channel is empty.
func (c Chan) TryReceive() bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

// ReadChan is a receive-only notification channel
type ReadChan <-chan struct{}

// Receive receive a notification from channel. It blocks if channel is empty.
func (c ReadChan) Receive() {
	<-c
}

// Receive receive a notification from channel. It returns false if channel is empty.
func (c ReadChan) TryReceive() bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

// WriteChan is a send-only notification channel
type WriteChan chan<- struct{}

// Send send a notification to channel. It blocks if channel is full.
func (c WriteChan) Send() {
	c <- Empty
}

// Send send a notification to channel. It returns false if channel is full.
func (c WriteChan) TrySend() bool {
	select {
	case c <- Empty:
		return true
	default:
		return false
	}
}

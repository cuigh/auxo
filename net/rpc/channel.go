package rpc

import (
	"net"

	"bufio"

	"github.com/cuigh/auxo/data/guid"
	"github.com/cuigh/auxo/security"
)

type Channel struct {
	net.Conn
	id string
	r  *bufio.Reader
	w  *bufio.Writer
	u  security.User
	//d data.Map
}

func newChannel(conn net.Conn) *Channel {
	return &Channel{
		id:   guid.New().String(),
		Conn: conn,
		r:    bufio.NewReader(conn),
		w:    bufio.NewWriter(conn),
	}
}

func (c *Channel) ID() string {
	return c.id
}

func (c *Channel) Reader() *bufio.Reader {
	return c.r
}

func (c *Channel) Writer() *bufio.Writer {
	return c.w
}

func (c *Channel) Peek(n int) ([]byte, error) {
	return c.r.Peek(n)
}

func (c *Channel) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *Channel) ReadByte() (byte, error) {
	return c.r.ReadByte()
}

func (c *Channel) ReadBytes(delim byte) ([]byte, error) {
	return c.r.ReadBytes(delim)
}

func (c *Channel) ReadString(delim byte) (string, error) {
	return c.r.ReadString(delim)
}

func (c *Channel) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *Channel) WriteByte(b byte) error {
	return c.w.WriteByte(b)
}

func (c *Channel) WriteString(s string) (int, error) {
	return c.w.WriteString(s)
}

func (c *Channel) Flush() error {
	return c.w.Flush()
}

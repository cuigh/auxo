package gsd

import "context"

type ExecuteInfo struct {
	query string
	args  []interface{}
}

type executeContext struct {
	sql  string
	args []interface{}
	db   *database
	Executor
	context.Context
}

func (c *executeContext) Reset() {
	c.sql = ""
	c.args = nil
}

func (c *executeContext) Value() (v Value) {
	v.err = c.Scan(&v.bytes)
	return
}

func (c *executeContext) Result() (ExecuteResult, error) {
	return c.Exec(c.Context, c.sql, c.args...)
}

func (c *executeContext) Scan(dst ...interface{}) error {
	return c.QueryRow(c.Context, c.sql, c.args...).Scan(dst...)
}

func (c *executeContext) Fill(i interface{}) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	return r.Fill(i)
}

func (c *executeContext) Reader() (Reader, error) {
	rows, err := c.QueryRows(c.Context, c.sql, c.args...)
	return (*reader)(rows), err
}

func (c *executeContext) For(fn func(r Reader) error) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}

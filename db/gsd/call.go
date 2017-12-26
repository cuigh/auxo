package gsd

import (
	"context"
	"database/sql"
)

type CallInfo struct {
	SP   string
	Args []interface{}
}

type callContext struct {
	Builder
	info CallInfo
	db   *database
	Executor
	context.Context
}

func (c *callContext) Reset() {
	c.info.SP = ""
	c.info.Args = nil
	c.Builder.Reset()
}

func (c *callContext) Call(sp string, args ...interface{}) ExecuteResult {
	c.info.SP = sp
	c.info.Args = args
	return c
}

func (c *callContext) Value() (v *Value) {
	var row *sql.Row
	if row, v.err = c.row(); v.err == nil {
		v.err = row.Scan(&v.bytes)
	}
	return
}

func (c *callContext) RowsAffected() (int64, error) {
	return 0, nil
}

func (c *callContext) LastInsertId() (int64, error) {
	return 0, nil
}

func (c *callContext) result() (sql.Result, error) {
	defer ctxPool.PutCall(c)

	err := c.db.p.BuildCall(&c.Builder, &c.info)
	if err != nil {
		return nil, err
	}
	return c.Exec(c.Context, c.Builder.String(), c.Builder.Args...)
}

func (c *callContext) row() (*sql.Row, error) {
	defer ctxPool.PutCall(c)

	err := c.db.p.BuildCall(&c.Builder, &c.info)
	if err != nil {
		return nil, err
	}
	return c.QueryRow(c.Context, c.Builder.String(), c.Builder.Args...), nil
}

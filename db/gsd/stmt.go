package gsd

import (
	"database/sql"
	"sync"
)

type stmtMap struct {
	locker sync.Mutex
	db     *sql.DB
	m      map[string]*Stmt
}

func newStmtMap(db *sql.DB) *stmtMap {
	return &stmtMap{
		db: db,
		m:  make(map[string]*Stmt),
	}
}

func (m *stmtMap) Get(name string, builder func() string) (stmt *Stmt, err error) {
	m.locker.Lock()
	stmt = m.m[name]
	if stmt == nil {
		var s *sql.Stmt
		if s, err = m.db.Prepare(builder()); err == nil {
			m.m[name] = (*Stmt)(s)
		}
	}
	m.locker.Unlock()
	return
}

type Stmt sql.Stmt

func (s *Stmt) Execute(args ...interface{}) ExecuteClause {
	return &stmtContext{
		stmt: (*sql.Stmt)(s),
		args: args,
	}
}

// Invoke execute stmt with named args, i must be map or struct
//func (s *Stmt) Invoke(args interface{}) ExecuteClause {
//	return &stmtContext{
//		stmt: (*sql.Stmt)(s),
//		args: args,
//	}
//}

type stmtContext struct {
	stmt *sql.Stmt
	args []interface{}
}

func (c *stmtContext) Value() (v Value) {
	v.err = c.Scan(&v.bytes)
	return
}

func (c *stmtContext) Result() (ExecuteResult, error) {
	return c.stmt.Exec(c.args...)
}

func (c *stmtContext) Scan(dst ...interface{}) error {
	return c.stmt.QueryRow(c.args...).Scan(dst...)
}

func (c *stmtContext) Fill(i interface{}) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	return r.Fill(i)
}

func (c *stmtContext) For(fn func(r Reader) error) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}

func (c *stmtContext) Reader() (Reader, error) {
	rows, err := c.stmt.Query(c.args...)
	return (*reader)(rows), err
}

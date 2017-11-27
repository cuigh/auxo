package gsd

import (
	"database/sql"

	"time"

	"github.com/cuigh/auxo/errors"
)

var ErrTXCancelled = errors.New("gsd: transaction has been cancelled")

type TX interface {
	Insert(table string) InsertClause
	Create(i interface{}, filter ...ColumnFilter) error
	Delete(table string) DeleteClause
	Remove(i interface{}) ResultClause
	Update(table string) UpdateClause
	Modify(i interface{}, filter ...ColumnFilter) ResultClause
	Select(cols ...string) SelectClause
	Query(cols *Columns, distinct ...bool) SelectClause
	Load(i interface{}) error
	Count(table interface{}) CountClause
	Execute(sql string, args ...interface{}) ExecuteClause
	Prepare(query string) (*Stmt, error)
	Stmt(name string, b func() string) (*Stmt, error)
	Commit() error
	Rollback() error
}

type transaction struct {
	db *database
	tx *sql.Tx
}

func (t *transaction) exec(query string, args ...interface{}) (sql.Result, error) {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.Exec(query, args...)
}

func (t *transaction) query(query string, args ...interface{}) (*sql.Rows, error) {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.Query(query, args...)
}

func (t *transaction) queryRow(query string, args ...interface{}) *sql.Row {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.QueryRow(query, args...)
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *transaction) Insert(table string) InsertClause {
	return ctxPool.GetInsert(t.db, t).Insert(table)
}

func (t *transaction) Create(i interface{}, filter ...ColumnFilter) error {
	return ctxPool.GetInsert(t.db, t).Create(i, filter...)
}

func (t *transaction) Delete(table string) DeleteClause {
	return ctxPool.GetDelete(t.db, t).Delete(table)
}

func (t *transaction) Remove(i interface{}) ResultClause {
	return ctxPool.GetDelete(t.db, t).Remove(i)
}

func (t *transaction) Update(table string) UpdateClause {
	return ctxPool.GetUpdate(t.db, t).Update(table)
}

func (t *transaction) Modify(i interface{}, filter ...ColumnFilter) ResultClause {
	return ctxPool.GetUpdate(t.db, t).Modify(i, filter...)
}

func (t *transaction) Select(cols ...string) SelectClause {
	return ctxPool.GetSelect(t.db, t).Select(NewColumns(cols...))
}

func (t *transaction) Query(cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(t.db, t).Select(cols, distinct...)
}

func (t *transaction) Load(i interface{}) error {
	return ctxPool.GetSelect(t.db, t).Load(i)
}

func (t *transaction) Count(table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(t.db, t)).Count(table)
}

func (t *transaction) Execute(sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: t.db, executor: t}
}

func (t *transaction) Prepare(query string) (*Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return (*Stmt)(stmt), nil
}

func (t *transaction) Stmt(name string, b func() string) (*Stmt, error) {
	return t.db.Stmt(name, b)
}

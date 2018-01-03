package gsd

import (
	"context"
	"database/sql"

	"time"

	"github.com/cuigh/auxo/errors"
)

var ErrTXCancelled = errors.New("gsd: transaction has been cancelled")

type TX interface {
	Insert(table string) InsertClause
	InsertContext(ctx context.Context, table string) InsertClause
	Create(i interface{}, filter ...ColumnFilter) error
	CreateContext(ctx context.Context, i interface{}, filter ...ColumnFilter) error
	Delete(table string) DeleteClause
	DeleteContext(ctx context.Context, table string) DeleteClause
	Remove(i interface{}) (r Result, err error)
	RemoveContext(ctx context.Context, i interface{}) (r Result, err error)
	Update(table string) UpdateClause
	UpdateContext(ctx context.Context, table string) UpdateClause
	Modify(i interface{}, filter ...ColumnFilter) (r Result, err error)
	ModifyContext(ctx context.Context, i interface{}, filter ...ColumnFilter) (r Result, err error)
	Select(cols ...string) SelectClause
	SelectContext(ctx context.Context, cols ...string) SelectClause
	Query(cols *Columns, distinct ...bool) SelectClause
	QueryContext(ctx context.Context, cols *Columns, distinct ...bool) SelectClause
	Load(i interface{}) error
	LoadContext(ctx context.Context, i interface{}) error
	Count(table interface{}) CountClause
	CountContext(ctx context.Context, table interface{}) CountClause
	Execute(sql string, args ...interface{}) ExecuteClause
	ExecuteContext(ctx context.Context, sql string, args ...interface{}) ExecuteClause
	Prepare(query string) (*Stmt, error)
	PrepareContext(ctx context.Context, query string) (*Stmt, error)
	Stmt(name string, b func() string) (*Stmt, error)
	Commit() error
	Rollback() error
}

type transaction struct {
	db *database
	tx *sql.Tx
	e  Executor
}

func (t *transaction) Database() string {
	return t.db.Database()
}

func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *transaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *transaction) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if t.db.opts.Trace.Enabled {
		defer t.db.trace(query, args, time.Now())
	}
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *transaction) Insert(table string) InsertClause {
	return ctxPool.GetInsert(context.Background(), t.db, t.e).Insert(table)
}

func (t *transaction) InsertContext(ctx context.Context, table string) InsertClause {
	return ctxPool.GetInsert(ctx, t.db, t.e).Insert(table)
}

func (t *transaction) Create(i interface{}, filter ...ColumnFilter) error {
	return ctxPool.GetInsert(context.Background(), t.db, t.e).Create(i, filter...)
}

func (t *transaction) CreateContext(ctx context.Context, i interface{}, filter ...ColumnFilter) error {
	return ctxPool.GetInsert(ctx, t.db, t.e).Create(i, filter...)
}

func (t *transaction) Delete(table string) DeleteClause {
	return ctxPool.GetDelete(context.Background(), t.db, t.e).Delete(table)
}

func (t *transaction) DeleteContext(ctx context.Context, table string) DeleteClause {
	return ctxPool.GetDelete(ctx, t.db, t.e).Delete(table)
}

func (t *transaction) Remove(i interface{}) (r Result, err error) {
	return ctxPool.GetDelete(context.Background(), t.db, t.e).Remove(i)
}

func (t *transaction) RemoveContext(ctx context.Context, i interface{}) (r Result, err error) {
	return ctxPool.GetDelete(ctx, t.db, t.e).Remove(i)
}

func (t *transaction) Update(table string) UpdateClause {
	return ctxPool.GetUpdate(context.Background(), t.db, t.e).Update(table)
}

func (t *transaction) UpdateContext(ctx context.Context, table string) UpdateClause {
	return ctxPool.GetUpdate(ctx, t.db, t.e).Update(table)
}

func (t *transaction) Modify(i interface{}, filter ...ColumnFilter) (r Result, err error) {
	return ctxPool.GetUpdate(context.Background(), t.db, t.e).Modify(i, filter...)
}

func (t *transaction) ModifyContext(ctx context.Context, i interface{}, filter ...ColumnFilter) (r Result, err error) {
	return ctxPool.GetUpdate(ctx, t.db, t.e).Modify(i, filter...)
}

func (t *transaction) Select(cols ...string) SelectClause {
	return ctxPool.GetSelect(context.Background(), t.db, t.e).Select(NewColumns(cols...))
}

func (t *transaction) SelectContext(ctx context.Context, cols ...string) SelectClause {
	return ctxPool.GetSelect(ctx, t.db, t.e).Select(NewColumns(cols...))
}

func (t *transaction) Query(cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(context.Background(), t.db, t.e).Select(cols, distinct...)
}

func (t *transaction) QueryContext(ctx context.Context, cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(ctx, t.db, t.e).Select(cols, distinct...)
}

func (t *transaction) Load(i interface{}) error {
	return ctxPool.GetSelect(context.Background(), t.db, t.e).Load(i)
}

func (t *transaction) LoadContext(ctx context.Context, i interface{}) error {
	return ctxPool.GetSelect(ctx, t.db, t.e).Load(i)
}

func (t *transaction) Count(table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(context.Background(), t.db, t.e)).Count(table)
}

func (t *transaction) CountContext(ctx context.Context, table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(ctx, t.db, t.e)).Count(table)
}

func (t *transaction) Execute(sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: t.db, Executor: t.e, Context: context.Background()}
}

func (t *transaction) ExecuteContext(ctx context.Context, sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: t.db, Executor: t.e, Context: ctx}
}

func (t *transaction) Prepare(query string) (*Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return (*Stmt)(stmt), nil
}

func (t *transaction) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return (*Stmt)(stmt), nil
}

func (t *transaction) Stmt(name string, b func() string) (*Stmt, error) {
	return t.db.Stmt(name, b)
}

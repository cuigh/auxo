package gsd

import (
	"context"
	"database/sql"
	"runtime/debug"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
)

type Options struct {
	Name         string
	Provider     string
	Driver       string
	Address      string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
	Trace        struct {
		Enabled bool
		Time    time.Duration
	}
	Options data.Map
}

type DB interface {
	Insert(ctx context.Context, table string) InsertClause
	Create(ctx context.Context, i interface{}, filter ...ColumnFilter) error
	Delete(ctx context.Context, table string) DeleteClause
	Remove(ctx context.Context, i interface{}) (r Result, err error)
	Update(ctx context.Context, table string) UpdateClause
	Modify(ctx context.Context, i interface{}, filter ...ColumnFilter) (r Result, err error)
	Select(ctx context.Context, cols ...string) SelectClause
	//Distinct(ctx context.Context, cols ...string) SelectClause
	Query(ctx context.Context, cols *Columns, distinct ...bool) SelectClause
	Load(ctx context.Context, i interface{}) error
	Count(ctx context.Context, table interface{}) CountClause
	Execute(ctx context.Context, sql string, args ...interface{}) ExecuteClause
	//Call(ctx context.Context, sp string, args ...interface{}) CallClause
	Prepare(ctx context.Context, query string) (*Stmt, error)
	Stmt(name string, b func() string) (*Stmt, error)
	Transact(ctx context.Context, fn func(tx TX) error, opts ...*sql.TxOptions) (err error)
}

type database struct {
	opts   *Options
	logger log.Logger
	db     *sql.DB
	p      Provider
	stmts  *stmtMap
	e      Executor
	Filters
}

func (d *database) Database() string {
	return d.opts.Name
}

func (d *database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.ExecContext(ctx, query, args...)
}

func (d *database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *database) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.QueryContext(ctx, query, args...)
}

func (d *database) Insert(ctx context.Context, table string) InsertClause {
	// must release ctx after execute
	return ctxPool.GetInsert(ctx, d, d.e).Insert(table)
}

func (d *database) Create(ctx context.Context, i interface{}, filter ...ColumnFilter) error { // (int64, error)
	return ctxPool.GetInsert(ctx, d, d.e).Create(i, filter...)
}

func (d *database) Delete(ctx context.Context, table string) DeleteClause {
	return ctxPool.GetDelete(ctx, d, d.e).Delete(table)
}

func (d *database) Remove(ctx context.Context, i interface{}) (r Result, err error) {
	return ctxPool.GetDelete(ctx, d, d.e).Remove(i)
}

func (d *database) Update(ctx context.Context, table string) UpdateClause {
	return ctxPool.GetUpdate(ctx, d, d.e).Update(table)
}

func (d *database) Modify(ctx context.Context, i interface{}, filter ...ColumnFilter) (r Result, err error) {
	return ctxPool.GetUpdate(ctx, d, d.e).Modify(i, filter...)
}

// Save inserts or updates table
//func (d *database) Save(ctx context.Context, i interface{}) {
//}

func (d *database) Select(ctx context.Context, cols ...string) SelectClause {
	return ctxPool.GetSelect(ctx, d, d.e).Select(NewColumns(cols...))
}

func (d *database) Query(ctx context.Context, cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(ctx, d, d.e).Select(cols, distinct...)
}

//func (d *database) Find(ctx context.Context, i interface{}) {
//}

// Load fetch a single row by primary keys. It returns ErrNoRows if no record was found.
func (d *database) Load(ctx context.Context, i interface{}) error {
	return ctxPool.GetSelect(ctx, d, d.e).Load(i)
}

func (d *database) Count(ctx context.Context, table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(ctx, d, d.e)).Count(table)
}

func (d *database) Execute(ctx context.Context, sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: d, Executor: d.e, Context: ctx}
}

//func (d *database) Call(ctx context.Context,sp string, args ...interface{}) (*sql.Rows, error) {
//	return d.db.Query(sp, args...)
//}

//func (d *database) Invoke(sp string, m map[string]interface{}) (*sql.Rows, error) {
//	return d.db.Query(sql, args...)
//}

//func (d *database) Stats() sql.DBStats {
//	return d.db.Stats()
//}

func (d *database) Transact(ctx context.Context, fn func(tx TX) error, opts ...*sql.TxOptions) (err error) {
	var trans *sql.Tx
	if len(opts) == 0 {
		trans, err = d.db.BeginTx(ctx, nil)
	} else {
		trans, err = d.db.BeginTx(ctx, opts[0])
	}
	if err != nil {
		return
	}

	tx := &transaction{db: d, tx: trans}
	tx.e = d.Filters.Apply(tx)
	defer func() {
		if e := recover(); e != nil {
			err = errors.Convert(e)
			d.rollback(tx)
			d.logger.Errorf("gsd > Transact panic: %v, stack:\n%s", e, debug.Stack())
		}
	}()

	err = fn(tx)
	switch err {
	case nil:
		err = tx.Commit()
	case ErrTXCancelled:
		err = nil
		fallthrough
	default:
		d.rollback(tx)
	}
	return
}

func (d *database) Prepare(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := d.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return (*Stmt)(stmt), nil
}

func (d *database) Stmt(name string, b func() string) (*Stmt, error) {
	return d.stmts.Get(name, b)
}

func (d *database) trace(sql string, args []interface{}, start time.Time) {
	t := time.Since(start)
	if t >= d.opts.Trace.Time {
		d.logger.Debugf("gsd > sql: %v, args: %v(%v), time: %v", sql, args, len(args), t)
	}
}

func (d *database) rollback(tx *transaction) {
	if err := tx.Rollback(); err != nil {
		d.logger.Errorf("gsd > TX rollback failed: ", err)
	}
}

type LazyDB struct {
	name string
	db   DB
}

func Lazy(name string) *LazyDB {
	return &LazyDB{name: name}
}

func (l *LazyDB) Try() (db DB, err error) {
	if l.db == nil {
		// we don't use locker here, because method Open is already safe
		l.db, err = Open(l.name)
	}
	return l.db, err
}

func (l *LazyDB) Get() (db DB) {
	if l.db == nil {
		// we don't use locker here, because method MustOpen is already safe
		l.db = MustOpen(l.name)
	}
	return l.db
}

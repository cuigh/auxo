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
	//Distinct(cols ...string) SelectClause
	Query(cols *Columns, distinct ...bool) SelectClause
	QueryContext(ctx context.Context, cols *Columns, distinct ...bool) SelectClause
	Load(i interface{}) error
	LoadContext(ctx context.Context, i interface{}) error
	Count(table interface{}) CountClause
	CountContext(ctx context.Context, table interface{}) CountClause
	Execute(sql string, args ...interface{}) ExecuteClause
	ExecuteContext(ctx context.Context, sql string, args ...interface{}) ExecuteClause
	//Call(sp string, args ...interface{}) CallClause
	Prepare(query string) (*Stmt, error)
	PrepareContext(ctx context.Context, query string) (*Stmt, error)
	Stmt(name string, b func() string) (*Stmt, error)
	Transact(fn func(tx TX) error, opts ...*sql.TxOptions) (err error)
	TransactContext(ctx context.Context, fn func(tx TX) error, opts ...*sql.TxOptions) (err error)
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

func (d *database) Insert(table string) InsertClause {
	// must release ctx after execute
	return ctxPool.GetInsert(context.Background(), d, d.e).Insert(table)
}
func (d *database) InsertContext(ctx context.Context, table string) InsertClause {
	// must release ctx after execute
	return ctxPool.GetInsert(ctx, d, d.e).Insert(table)
}

func (d *database) Create(i interface{}, filter ...ColumnFilter) error { // (int64, error)
	return ctxPool.GetInsert(context.Background(), d, d.e).Create(i, filter...)
}

func (d *database) CreateContext(ctx context.Context, i interface{}, filter ...ColumnFilter) error { // (int64, error)
	return ctxPool.GetInsert(ctx, d, d.e).Create(i, filter...)
}

func (d *database) Delete(table string) DeleteClause {
	return ctxPool.GetDelete(context.Background(), d, d.e).Delete(table)
}

func (d *database) DeleteContext(ctx context.Context, table string) DeleteClause {
	return ctxPool.GetDelete(ctx, d, d.e).Delete(table)
}

func (d *database) Remove(i interface{}) (r Result, err error) {
	return ctxPool.GetDelete(context.Background(), d, d.e).Remove(i)
}
func (d *database) RemoveContext(ctx context.Context, i interface{}) (r Result, err error) {
	return ctxPool.GetDelete(ctx, d, d.e).Remove(i)
}

func (d *database) Update(table string) UpdateClause {
	return ctxPool.GetUpdate(context.Background(), d, d.e).Update(table)
}

func (d *database) UpdateContext(ctx context.Context, table string) UpdateClause {
	return ctxPool.GetUpdate(ctx, d, d.e).Update(table)
}

func (d *database) Modify(i interface{}, filter ...ColumnFilter) (r Result, err error) {
	return ctxPool.GetUpdate(context.Background(), d, d.e).Modify(i, filter...)
}
func (d *database) ModifyContext(ctx context.Context, i interface{}, filter ...ColumnFilter) (r Result, err error) {
	return ctxPool.GetUpdate(ctx, d, d.e).Modify(i, filter...)
}

// Save inserts or updates table
//func (d *database) Save(i interface{}) {
//}

func (d *database) Select(cols ...string) SelectClause {
	return ctxPool.GetSelect(context.Background(), d, d.e).Select(NewColumns(cols...))
}
func (d *database) SelectContext(ctx context.Context, cols ...string) SelectClause {
	return ctxPool.GetSelect(ctx, d, d.e).Select(NewColumns(cols...))
}

func (d *database) Query(cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(context.Background(), d, d.e).Select(cols, distinct...)
}

func (d *database) QueryContext(ctx context.Context, cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(ctx, d, d.e).Select(cols, distinct...)
}

//func (d *database) Find(i interface{}) {
//}

// Load fetch a single row by primary keys. It returns ErrNoRows if no record was found.
func (d *database) Load(i interface{}) error {
	return ctxPool.GetSelect(context.Background(), d, d.e).Load(i)
}

// Load fetch a single row by primary keys. It returns ErrNoRows if no record was found.
func (d *database) LoadContext(ctx context.Context, i interface{}) error {
	return ctxPool.GetSelect(ctx, d, d.e).Load(i)
}

func (d *database) Count(table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(context.Background(), d, d.e)).Count(table)
}

func (d *database) CountContext(ctx context.Context, table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(ctx, d, d.e)).Count(table)
}

func (d *database) Execute(sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: d, Executor: d.e, Context: context.Background()}
}

func (d *database) ExecuteContext(ctx context.Context, sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: d, Executor: d.e, Context: ctx}
}

//func (d *database) Call(sp string, args ...interface{}) (*sql.Rows, error) {
//	return d.db.Query(sp, args...)
//}
//
//func (d *database) CallContext(ctx context.Context,sp string, args ...interface{}) (*sql.Rows, error) {
//	return d.db.Query(sp, args...)
//}

//func (d *database) Invoke(sp string, m map[string]interface{}) (*sql.Rows, error) {
//	return d.db.Query(sql, args...)
//}

//func (d *database) Stats() sql.DBStats {
//	return d.db.Stats()
//}

func (d *database) Transact(fn func(tx TX) error, opts ...*sql.TxOptions) (err error) {
	return d.TransactContext(context.Background(), fn, opts...)
}

func (d *database) TransactContext(ctx context.Context, fn func(tx TX) error, opts ...*sql.TxOptions) (err error) {
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

func (d *database) Prepare(query string) (*Stmt, error) {
	stmt, err := d.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	return (*Stmt)(stmt), nil
}

func (d *database) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
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

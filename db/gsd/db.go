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
	Create(i interface{}, filter ...ColumnFilter) error
	Delete(table string) DeleteClause
	Remove(i interface{}) ResultClause
	Update(table string) UpdateClause
	Modify(i interface{}, filter ...ColumnFilter) ResultClause
	Select(cols ...string) SelectClause
	//Distinct(cols ...string) SelectClause
	Query(cols *Columns, distinct ...bool) SelectClause
	Load(i interface{}) error
	Count(table interface{}) CountClause
	Execute(sql string, args ...interface{}) ExecuteClause
	//Call(sp string, args ...interface{}) CallClause
	Prepare(query string) (*Stmt, error)
	Stmt(name string, b func() string) (*Stmt, error)
	Transact(fn func(tx TX) error, opts ...*sql.TxOptions) (err error)
}

type database struct {
	opts   *Options
	logger *log.Logger
	db     *sql.DB
	p      Provider
	stmts  *stmtMap
}

func (d *database) exec(query string, args ...interface{}) (sql.Result, error) {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.Exec(query, args...)
}

func (d *database) query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.Query(query, args...)
}

func (d *database) queryRow(query string, args ...interface{}) *sql.Row {
	if d.opts.Trace.Enabled {
		defer d.trace(query, args, time.Now())
	}
	return d.db.QueryRow(query, args...)
}

func (d *database) Insert(table string) InsertClause {
	// must release ctx after execute
	return ctxPool.GetInsert(d, d).Insert(table)
}

func (d *database) Create(i interface{}, filter ...ColumnFilter) error { // (int64, error)
	return ctxPool.GetInsert(d, d).Create(i, filter...)
}

func (d *database) Delete(table string) DeleteClause {
	return ctxPool.GetDelete(d, d).Delete(table)
}

func (d *database) Remove(i interface{}) ResultClause {
	return ctxPool.GetDelete(d, d).Remove(i)
}

func (d *database) Update(table string) UpdateClause {
	return ctxPool.GetUpdate(d, d).Update(table)
}

func (d *database) Modify(i interface{}, filter ...ColumnFilter) ResultClause {
	return ctxPool.GetUpdate(d, d).Modify(i, filter...)
}

// Save inserts or updates table
//func (d *database) Save(i interface{}) {
//}

func (d *database) Select(cols ...string) SelectClause {
	return ctxPool.GetSelect(d, d).Select(NewColumns(cols...))
}

func (d *database) Query(cols *Columns, distinct ...bool) SelectClause {
	return ctxPool.GetSelect(d, d).Select(cols, distinct...)
}

//func (d *database) Find(i interface{}) {
//}

// Load fetch a single row by primary keys. It returns ErrNoRows if no record was found.
func (d *database) Load(i interface{}) error {
	return ctxPool.GetSelect(d, d).Load(i)
}

func (d *database) Count(table interface{}) CountClause {
	return (*countContext)(ctxPool.GetSelect(d, d)).Count(table)
}

func (d *database) Execute(sql string, args ...interface{}) ExecuteClause {
	return &executeContext{sql: sql, args: args, db: d, executor: d}
}

func (d *database) Call(sp string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(sp, args...)
}

//func (d *database) Invoke(sp string, m map[string]interface{}) (*sql.Rows, error) {
//	return d.db.Query(sql, args...)
//}

//func (d *database) Stats() sql.DBStats {
//	return d.db.Stats()
//}

func (d *database) Transact(fn func(tx TX) error, opts ...*sql.TxOptions) (err error) {
	var trans *sql.Tx
	if len(opts) == 0 {
		trans, err = d.db.Begin()
	} else {
		trans, err = d.db.BeginTx(context.TODO(), opts[0])
	}
	if err != nil {
		return
	}

	tx := &transaction{db: d, tx: trans}
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

package gsd

import (
	"database/sql"
	"sync"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
)

const PkgName = "auxo.db.gsd"

var (
	ErrNoRows = sql.ErrNoRows
)

var (
	factory   = &Factory{dbMap: make(map[string]DB)}
	providers = make(map[string]ProviderBuilder)
	ctxPool   = newContextPool()
)

type ProviderBuilder func(options data.Map) Provider

type Provider interface {
	BuildInsert(b *Builder, ctx *InsertInfo) (err error)
	BuildDelete(b *Builder, ctx *DeleteInfo) (err error)
	BuildUpdate(b *Builder, ctx *UpdateInfo) (err error)
	BuildSelect(b *Builder, ctx *SelectInfo) (err error)
	BuildCall(b *Builder, ctx *CallInfo) (err error)
}

func RegisterProvider(name string, builder ProviderBuilder) {
	providers[name] = builder
}

type contextPool struct {
	inserts sync.Pool
	deletes sync.Pool
	updates sync.Pool
	selects sync.Pool
	calls   sync.Pool
}

func newContextPool() *contextPool {
	p := &contextPool{}
	p.inserts = sync.Pool{
		New: func() interface{} {
			return &insertContext{}
		},
	}
	p.deletes = sync.Pool{
		New: func() interface{} {
			return &deleteContext{}
		},
	}
	p.updates = sync.Pool{
		New: func() interface{} {
			return &updateContext{}
		},
	}
	p.selects = sync.Pool{
		New: func() interface{} {
			return &selectContext{}
		},
	}
	p.calls = sync.Pool{
		New: func() interface{} {
			return &callContext{}
		},
	}
	return p
}

func (p *contextPool) GetInsert(db *database, q executor) (ctx *insertContext) {
	ctx = p.inserts.Get().(*insertContext)
	ctx.db, ctx.executor = db, q
	return
}

func (p *contextPool) PutInsert(c *insertContext) {
	c.Reset()
	p.inserts.Put(c)
}

func (p *contextPool) GetDelete(db *database, q executor) (ctx *deleteContext) {
	ctx = p.deletes.Get().(*deleteContext)
	ctx.db, ctx.executor = db, q
	return ctx
}

func (p *contextPool) PutDelete(c *deleteContext) {
	c.Reset()
	p.deletes.Put(c)
}

func (p *contextPool) GetUpdate(db *database, q executor) (ctx *updateContext) {
	ctx = p.updates.Get().(*updateContext)
	ctx.db, ctx.executor = db, q
	return ctx
}

func (p *contextPool) PutUpdate(c *updateContext) {
	c.Reset()
	p.updates.Put(c)
}

func (p *contextPool) GetSelect(db *database, q executor) (ctx *selectContext) {
	ctx = p.selects.Get().(*selectContext)
	ctx.db, ctx.executor = db, q
	return ctx
}

func (p *contextPool) PutSelect(c *selectContext) {
	c.Reset()
	p.selects.Put(c)
}

func (p *contextPool) GetCall(db *database, q executor) (ctx *callContext) {
	ctx = p.calls.Get().(*callContext)
	ctx.db, ctx.executor = db, q
	return ctx
}

func (p *contextPool) PutCall(c *callContext) {
	c.Reset()
	p.selects.Put(c)
}

type Builder struct {
	texts.Builder
	Args []interface{}
}

func (b *Builder) Reset() {
	b.Builder.Reset()
	if b.Args != nil {
		b.Args = b.Args[:0] //nil
	}
}

type executor interface {
	exec(query string, args ...interface{}) (sql.Result, error)
	query(query string, args ...interface{}) (*sql.Rows, error)
	queryRow(query string, args ...interface{}) *sql.Row
}

type Factory struct {
	locker sync.Mutex
	dbMap  map[string]DB
}

func New(opts *Options) (DB, error) {
	if opts.Address == "" {
		return nil, errors.New("gsd: New with empty address")
	}

	pb := providers[opts.Provider]
	if pb == nil {
		return nil, errors.New("gsd: can't find provider: " + opts.Provider)
	}

	driver := opts.Driver
	if driver == "" {
		driver = opts.Provider
	}
	db, err := sql.Open(driver, opts.Address)
	if err != nil {
		return nil, err
	}

	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnLifetime)
	}

	return &database{
		logger: log.Get(PkgName),
		opts:   opts,
		p:      pb(opts.Options),
		db:     db,
		stmts:  newStmtMap(db),
	}, nil
}

func Open(name string) (db DB, err error) {
	return factory.Open(name)
}

func MustOpen(name string) (db DB) {
	db, err := factory.Open(name)
	if err == nil {
		return db
	}
	panic(err)
}

func (f *Factory) Open(name string) (db DB, err error) {
	db = f.dbMap[name]
	if db == nil {
		db, err = f.initDB(name)
	}
	return
}

func (f *Factory) initDB(name string) (db DB, err error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	db = f.dbMap[name]
	if db != nil {
		return db, nil
	}

	opts, err := f.loadOptions(name)
	if err != nil {
		return nil, err
	}

	// create new map to avoid locking
	if db, err = New(opts); err == nil {
		dbs := make(map[string]DB)
		for k, v := range f.dbMap {
			dbs[k] = v
		}
		dbs[name] = db
		f.dbMap = dbs
	}
	return
}

func (f *Factory) loadOptions(name string) (*Options, error) {
	key := "db.sql." + name
	if config.Get(key) == nil {
		return nil, errors.Format("can't find db config for [%s]", name)
	}

	opts := &Options{}
	err := config.UnmarshalOption(key, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

package mongo

import (
	"sync"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	PkgName = "auxo.db.mongo"
)

var f = &factory{
	sessions: make(map[string]*Session),
}

type Database = mgo.Database
type Session = mgo.Session
type Collection = mgo.Collection
type GridFS = mgo.GridFS
type M = bson.M
type Query = mgo.Query
type Iter = mgo.Iter
type ChangeInfo = mgo.ChangeInfo
type MapReduceInfo = mgo.MapReduceInfo

func Configure(name string, opts *Options) {
	f.opts.Store(name, opts)
}

func Open(name string) (DB, error) {
	return f.Open(name)
}

func MustOpen(name string) DB {
	db, err := f.Open(name)
	if err == nil {
		return db
	}
	panic(err)
}

func With(name string, fn func(db DB) error) error {
	db, err := f.Open(name)
	if err != nil {
		return err
	}
	defer db.Close()

	return fn(db)
}

type Options struct {
	Address     string
	PoolSize    int
	DialTimeout time.Duration
	Consistency string
}

type DB interface {
	//Session() *Session
	C(name string) *Collection
	FS(prefix string) *GridFS
	Close()
}

type database struct {
	db *Database
}

// Session return original Session hold by this database.
//func (d *database) Session() *Session {
//	return d.db.Session
//}

func (d *database) C(name string) *Collection {
	return d.db.C(name)
}

func (d *database) FS(prefix string) *GridFS {
	return d.db.GridFS(prefix)
}

func (d *database) Close() {
	d.db.Session.Close()
}

type factory struct {
	locker   sync.Mutex
	sessions map[string]*Session
	opts     sync.Map
}

func (f *factory) Open(name string) (DB, error) {
	session := f.sessions[name]
	if session == nil {
		var err error
		session, err = f.initSession(name)
		if err != nil {
			return nil, err
		}
	}

	return &database{
		db: session.Copy().DB(""),
	}, nil
}

func (f *factory) initSession(name string) (*Session, error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	// check again
	p := f.sessions[name]
	if p != nil {
		return p, nil
	}

	var opts *Options
	if v, ok := f.opts.Load(name); ok {
		opts = v.(*Options)
	}
	if opts == nil {
		opts = &Options{}
		err := config.UnmarshalOption("db.mongo."+name, opts)
		if err != nil {
			return nil, err
		}
	}

	session, err := f.openSession(opts)
	if err == nil {
		// rebuild map to avoid locker
		sessions := make(map[string]*Session)
		for k, v := range f.sessions {
			sessions[k] = v
		}
		sessions[name] = p
		f.sessions = sessions
	}
	return session, err
}

func (f *factory) openSession(opts *Options) (*Session, error) {
	info, err := mgo.ParseURL(opts.Address)
	if err != nil {
		return nil, err
	}

	if opts.DialTimeout > 0 {
		info.Timeout = opts.DialTimeout
	}
	if opts.PoolSize > 0 {
		info.PoolLimit = opts.PoolSize
	}

	s, err := mgo.DialWithInfo(info)
	if err == nil && opts.Consistency != "" {
		switch opts.Consistency {
		case "Primary":
			s.SetMode(mgo.Primary, false)
		case "PrimaryPreferred":
			s.SetMode(mgo.PrimaryPreferred, false)
		case "Secondary":
			s.SetMode(mgo.Secondary, false)
		case "SecondaryPreferred":
			s.SetMode(mgo.SecondaryPreferred, false)
		case "Nearest":
			s.SetMode(mgo.Nearest, false)
		}
	}
	return s, err
}

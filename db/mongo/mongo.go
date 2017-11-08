package mongo

import (
	"sync"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/times"
	"github.com/cuigh/auxo/util/cast"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	PkgName         = "auxo.db.mongo"
	defaultPoolSize = 100
)

var factory = &Factory{
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
	factory.opts.Store(name, opts)
}

func Open(name string) (DB, error) {
	return factory.Open(name)
}

func MustOpen(name string) DB {
	db, err := factory.Open(name)
	if err == nil {
		return db
	}
	panic(err)
}

type Options struct {
	URL     string `option:"url"`
	Options data.Map
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

type Factory struct {
	locker   sync.Mutex
	sessions map[string]*Session
	opts     sync.Map
}

func (f *Factory) Open(name string) (DB, error) {
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

func (f *Factory) initSession(name string) (*Session, error) {
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
		pools := make(map[string]*Session)
		for k, v := range f.sessions {
			pools[k] = v
		}
		pools[name] = p
	}
	return session, err
}

func (f *Factory) openSession(opts *Options) (*Session, error) {
	maxPoolSize := cast.ToInt(opts.Options.Get("MaxPoolSize"))
	if maxPoolSize <= 0 {
		maxPoolSize = defaultPoolSize
	}

	info, err := mgo.ParseURL(opts.URL)
	if err != nil {
		return nil, err
	}

	info.Timeout = cast.ToDuration(opts.Options.Get("ConnectTimeout"), times.Seconds(10))
	info.PoolLimit = maxPoolSize

	s, err := mgo.DialWithInfo(info)
	if err == nil {
		if consistency := cast.ToString(opts.Options.Get("ReadPreference")); consistency != "" {
			switch consistency {
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
	}
	return s, err
}

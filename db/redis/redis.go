package redis

import (
	"sync"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
	"github.com/go-redis/redis"
)

const PkgName = "auxo.db.redis"

const (
	//TypeSingle   = "single"
	TypeRing     = "ring"
	TypeSentinel = "sentinel"
	TypeCluster  = "cluster"
)

var (
	Nil = redis.Nil
	f   = new(factory)
)

type Client = redis.Cmdable
type Pipeliner = redis.Pipeliner
type Cmd = redis.Cmd
type StringCmd = redis.StringCmd
type IntCmd = redis.IntCmd
type BoolCmd = redis.BoolCmd
type DurationCmd = redis.DurationCmd

func Open(name string) (Client, error) {
	return f.Open(name)
}

type factory struct {
	locker sync.Mutex
	cmds   map[string]Client
}

type Options struct {
	Type         string
	Address      []string
	Password     string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MasterNames  []string
	Db           int
}

func (f *factory) Open(name string) (cmd Client, err error) {
	cmd = f.cmds[name]
	if cmd == nil {
		cmd, err = f.create(name)
	}
	return
}

func (f *factory) create(name string) (cmd Client, err error) {
	f.locker.Lock()
	defer f.locker.Unlock()

	cmd = f.cmds[name]
	if cmd != nil {
		return cmd, nil
	}

	opts, err := f.loadOptions(name)
	if opts == nil {
		return nil, err
	}

	switch opts.Type {
	case TypeRing:
		cmd = f.createRing(opts)
	case TypeSentinel:
		cmd = f.createSentinel(opts)
	case TypeCluster:
		cmd = f.createCluster(opts)
	default: // single node
		cmd = f.createSingle(opts)
	}

	// rebuild map to avoid locker
	cmds := make(map[string]Client)
	for k, v := range f.cmds {
		cmds[k] = v
	}
	cmds[name] = cmd
	f.cmds = cmds
	return
}

func (f *factory) createRing(opts *Options) Client {
	ringOpts := &redis.RingOptions{
		Password:     opts.Password,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
	}
	ringOpts.Addrs = make(map[string]string)
	for _, addr := range opts.Address {
		ringOpts.Addrs[addr] = addr
	}
	return redis.NewRing(ringOpts)
}

func (f *factory) createSentinel(opts *Options) Client {
	if len(opts.MasterNames) > 1 {
		// TODO: support sentinel cluster
		panic(errors.NotSupported)
		// switch to Ring when having multiple MasterNames
		//ring := redis.NewRing(&redis.RingOptions{})
		//for _, master := range masterNames {
		//	options := &redis.FailoverOptions{
		//		MasterName:    master,
		//		SentinelAddrs: opts.Address,
		//		Password:      cast.ToString(opts.Options.Get("password")),
		//		DialTimeout:   cast.ToDuration(opts.Options.Get("connect_timeout")),
		//		ReadTimeout:   cast.ToDuration(opts.Options.Get("read_timeout")),
		//		WriteTimeout:  cast.ToDuration(opts.Options.Get("write_timeout")),
		//		PoolSize:      cast.ToInt(opts.Options.Get("max_pool_size")),
		//	}
		//	ring.addClient(master, redis.NewFailoverClient(options))
		//}
		//return ring
	} else {
		options := &redis.FailoverOptions{
			MasterName:    opts.MasterNames[0],
			SentinelAddrs: opts.Address,
			Password:      opts.Password,
			DialTimeout:   opts.DialTimeout,
			ReadTimeout:   opts.ReadTimeout,
			WriteTimeout:  opts.WriteTimeout,
			PoolSize:      opts.PoolSize,
		}
		return redis.NewFailoverClient(options)
	}
}

func (f *factory) createCluster(opts *Options) Client {
	options := &redis.ClusterOptions{
		Addrs:        opts.Address,
		Password:     opts.Password,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
	}
	return redis.NewClusterClient(options)
}

func (f *factory) createSingle(opts *Options) Client {
	options := &redis.Options{
		Addr:         opts.Address[0],
		Password:     opts.Password,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		DB:           opts.Db,
	}
	return redis.NewClient(options)
}

func (f *factory) loadOptions(name string) (*Options, error) {
	key := "db.redis." + name
	if !config.Exist(key) {
		return nil, errors.Format("can't find redis config for [%s]", name)
	}

	opts := &Options{}
	err := config.UnmarshalOption(key, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

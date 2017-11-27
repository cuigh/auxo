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

var factory = new(Factory)

func Open(name string) (redis.Cmdable, error) {
	return factory.Open(name)
}

type Factory struct {
	locker sync.Mutex
	cmds   map[string]redis.Cmdable
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
}

func (f *Factory) Open(name string) (cmd redis.Cmdable, err error) {
	cmd = f.cmds[name]
	if cmd == nil {
		cmd, err = f.create(name)
	}
	return
}

func (f *Factory) create(name string) (cmd redis.Cmdable, err error) {
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
	return
}

func (f *Factory) createRing(opts *Options) redis.Cmdable {
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

func (f *Factory) createSentinel(opts *Options) redis.Cmdable {
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

func (f *Factory) createCluster(opts *Options) redis.Cmdable {
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

func (f *Factory) createSingle(opts *Options) redis.Cmdable {
	options := &redis.Options{
		Addr:         opts.Address[0],
		Password:     opts.Password,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
	}
	return redis.NewClient(options)
}

func (f *Factory) loadOptions(name string) (*Options, error) {
	key := "db.redis." + name
	if config.Get(key) == nil {
		return nil, errors.Format("can't find redis config for [%s]", name)
	}

	opts := &Options{}
	err := config.UnmarshalOption(key, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

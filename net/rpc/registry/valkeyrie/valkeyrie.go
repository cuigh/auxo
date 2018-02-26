package valkeyrie

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/abronan/valkeyrie"
	"github.com/abronan/valkeyrie/store"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/data/set"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/util/cast"
	"github.com/cuigh/auxo/util/retry"
	"github.com/cuigh/auxo/util/run"
)

const PkgName = "auxo.net.rpc.registry.valkeyrie"

// Builder implements a common Builder based on valkeyrie.
type Builder struct {
	Backend store.Backend
}

func (b *Builder) Name() string {
	return string(b.Backend)
}

func (b *Builder) Build(server registry.Server, opts data.Map) (registry.Registry, error) {
	addrs := strings.Split(opts.Get("address").(string), ",")
	timeout := cast.ToDuration(opts.Get("dial_timeout"), 10*time.Second)
	username := cast.ToString(opts.Get("username"))
	password := cast.ToString(opts.Get("password"))
	interval := cast.ToDuration(opts.Get("heartbeat_interval"))
	if interval <= 0 {
		interval = 30 * time.Second
	}
	prefix := cast.ToString(opts.Get("prefix"))
	if prefix == "" {
		prefix = "/auxo/app"
	}

	// create store
	kv, err := valkeyrie.NewStore(b.Backend, addrs, &store.Config{
		ConnectionTimeout: timeout,
		Username:          username,
		Password:          password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store")
	}

	return &Registry{
		server:   server,
		store:    kv,
		key:      prefix + "/" + server.Name,
		interval: interval,
		ttl:      interval + 5*time.Second,
		logger:   log.Get(PkgName),
	}, nil
}

// Builder implements a common Registry based on valkeyrie.
type Registry struct {
	server   registry.Server
	store    store.Store
	key      string
	interval time.Duration
	ttl      time.Duration
	logger   log.Logger
	canceler run.Canceler
}

func (r *Registry) Register() {
	if r.canceler == nil {
		r.register()
		r.canceler = run.Schedule(r.interval, r.register, nil)
	}
}

func (r *Registry) register() {
	for _, addr := range r.server.Addresses {
		key := r.key + "/nodes/" + addr.URL
		m := data.Map{"version": r.server.Version}
		if r.server.Options != nil {
			m.Merge(r.server.Options)
		}
		m.Merge(addr.Options)
		b, err := json.Marshal(m)
		if err != nil {
			r.logger.Errorf("valkeyrie > Failed to marshal options of address '%s': %v", addr.URL, err)
			continue
		}

		err = r.store.Put(key, b, &store.WriteOptions{TTL: r.ttl})
		if err != nil {
			r.logger.Errorf("valkeyrie > Failed to register address '%s': %v", addr.URL, err)
		} else {
			r.logger.Debugf("valkeyrie > Register address '%s' successfully", addr.URL)
		}
	}
}

func (r *Registry) Offline() error {
	return retry.Do(3, nil, func() error {
		key := r.key + "/options"
		pair, opts, err := r.getOptions(key)
		if err != nil {
			return err
		}

		return r.updateOptions(key, pair, r.offline(opts))
	})
}

func (r *Registry) Online() error {
	return retry.Do(3, nil, func() error {
		key := r.key + "/options"
		pair, opts, err := r.getOptions(key)
		if err != nil {
			return err
		}

		return r.updateOptions(key, pair, r.online(opts))
	})
}

func (r *Registry) Close() {
	if r.canceler != nil {
		r.canceler.Cancel()
		r.logger.Debug("valkeyrie > Registry stopped")
	}
	r.store.Close()
}

func (r *Registry) getOptions(key string) (pair *store.KVPair, opts data.Map, err error) {
	// $key/options={"groups": {"test", ["192.168.50.150:9999"]}, "offline_nodes": ["192.168.50.151:9999"]}
	pair, err = r.store.Get(key, nil)
	if err == nil {
		opts = make(data.Map)
		err = json.Unmarshal(pair.Value, &opts)
	} else if err == store.ErrKeyNotFound {
		err = nil
	}
	return
}

func (r *Registry) updateOptions(key string, previous *store.KVPair, opts data.Map) (err error) {
	b, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	_, _, err = r.store.AtomicPut(key, b, previous, nil)
	return err
}

func (r *Registry) online(opts data.Map) data.Map {
	if opts == nil {
		return opts
	}

	v := opts.Get("offline_nodes")
	if v == nil {
		return opts
	}

	ss := set.NewStringSet(v.([]string)...)
	ss.RemoveSlice(r.server.Addresses, func(i int) string {
		return r.server.Addresses[i].URL
	})
	opts["offline_nodes"] = ss.Slice()
	return opts
}

func (r *Registry) offline(opts data.Map) data.Map {
	if opts == nil {
		opts = data.Map{}
	}

	nodes := make([]string, len(r.server.Addresses))
	for i, addr := range r.server.Addresses {
		nodes[i] = addr.URL
	}

	v := opts.Get("offline_nodes")
	if v == nil {
		opts["offline_nodes"] = nodes
		return opts
	}

	ss := set.NewStringSet(v.([]string)...)
	ss.Add(nodes...)
	opts["offline_nodes"] = ss.Slice()
	return opts
}

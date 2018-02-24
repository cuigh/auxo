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
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/transport"
	"github.com/cuigh/auxo/util/cast"
	"github.com/cuigh/auxo/util/run"
	"github.com/cuigh/auxo/util/semver"
)

const PkgName = "auxo.net.rpc.resolver.valkeyrie"

// Builder implements a common Builder based on valkeyrie.
type Builder struct {
	Backend store.Backend
}

func (b *Builder) Name() string {
	return string(b.Backend)
}

func (b *Builder) Build(client resolver.Client, opts data.Map) (resolver.Resolver, error) {
	addrs := strings.Split(opts.Get("address").(string), ",")
	timeout := cast.ToDuration(opts.Get("dial_timeout"), 10*time.Second)
	username := cast.ToString(opts.Get("username"))
	password := cast.ToString(opts.Get("password"))
	interval := cast.ToDuration(opts.Get("refresh_interval"))
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

	r := &Resolver{
		client:   client,
		store:    kv,
		key:      prefix + "/" + client.Server,
		interval: interval,
		logger:   log.Get(PkgName),
	}
	if client.Version != "" {
		r.constraint, err = semver.NewConstraint(client.Version)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Builder implements a common Resolver based on valkeyrie.
type Resolver struct {
	client     resolver.Client
	store      store.Store
	key        string
	interval   time.Duration
	logger     log.Logger
	constraint *semver.Constraints
	canceler   run.Canceler
}

func (r *Resolver) Resolve() ([]transport.Address, error) {
	return r.getAddrs()
}

func (r *Resolver) Watch(notify func([]transport.Address)) {
	if r.canceler == nil {
		r.canceler = run.Schedule(r.interval, func() {
			addrs, err := r.getAddrs()
			if err != nil {
				r.logger.Error("valkeyrie > Failed to refresh addresses: ", err)
			} else {
				notify(addrs)
			}
		}, nil)
	}
}

func (r *Resolver) Close() {
	if r.canceler != nil {
		r.canceler.Cancel()
		r.logger.Debug("valkeyrie > Resolver stopped")
	}
	r.store.Close()
}

func (r *Resolver) getAddrs() ([]transport.Address, error) {
	// $key/nodes/192.168.50.150:9999
	// $key/nodes/192.168.50.151:9999
	key := r.key + "/nodes"
	nodes, err := r.store.List(key, nil)
	if err == store.ErrKeyNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	opts, err := r.getOptions()
	if err != nil {
		return nil, err
	}

	addrs := make([]transport.Address, 0, len(nodes))
	for _, node := range nodes {
		addr := transport.Address{
			URL:     node.Key[len(key)+1:],
			Options: make(data.Map),
		}
		err = json.Unmarshal(node.Value, &addr.Options)
		if err != nil {
			return nil, err
		}
		if !r.filter(&addr, &opts) {
			addrs = append(addrs, addr)
		}
	}
	return addrs, nil
}

func (r *Resolver) getOptions() (opts appOptions, err error) {
	// $key/options={"groups": {"test", ["192.168.50.150:9999"]}, "offline_nodes": ["192.168.50.151:9999"]}
	var (
		key  = r.key + "/options"
		pair *store.KVPair
	)
	pair, err = r.store.Get(key, nil)
	if err == nil {
		err = json.Unmarshal(pair.Value, &opts)
	} else if err == store.ErrKeyNotFound {
		err = nil
	}
	return
}

func (r *Resolver) filter(addr *transport.Address, opts *appOptions) bool {
	if opts.offline(addr.URL) {
		r.logger.Debugf("valkeyrie > Drop offline node '%s'", addr.URL)
		return true
	}

	if r.client.Codec != "" {
		v := cast.ToString(addr.Options.Get("codec"))
		if v != "" {
			if s := set.NewStringSet(strings.Split(v, ",")...); !s.Contains(r.client.Codec) {
				return true
			}
		}
	}

	if r.constraint != nil {
		v := cast.ToString(addr.Options.Get("version"))
		if v == "" {
			r.logger.Debugf("valkeyrie > Drop node '%s' which does not have a version option", addr.URL)
			return true
		}

		ver, err := semver.NewVersion(v)
		if err != nil {
			r.logger.Debugf("valkeyrie > Drop node '%s(version: %s)' which has invalid version", addr.URL, v)
			return true
		}

		if !r.constraint.Check(ver) {
			r.logger.Debugf("valkeyrie > Drop node '%s(version: %s)' which does not meet version constraint", addr.URL, v)
			return true
		}
	}
	return false
}

type appOptions struct {
	//Groups       map[string][]string `json:"groups"`
	OfflineNodes []string `json:"offline_nodes"`
}

func (opts *appOptions) offline(addr string) bool {
	for _, n := range opts.OfflineNodes {
		if n == addr {
			return true
		}
	}
	return false
}

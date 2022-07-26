package dns

import (
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/transport"
	"github.com/cuigh/auxo/util/cast"
	"github.com/cuigh/auxo/util/run"
	"net"
	"time"
)

const PkgName = "auxo.net.rpc.resolver.dns"

func init() {
	resolver.Register(Builder{})
}

type Builder struct{}

func (Builder) Name() string {
	return "dns"
}

func (Builder) Build(_ resolver.Client, opts data.Map) (resolver.Resolver, error) {
	addr := cast.ToString(opts.Get("address"))
	interval := cast.ToDuration(opts.Get("refresh_interval"))
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return &Resolver{
		addr:     addr,
		interval: interval,
		logger:   log.Get(PkgName),
	}, nil
}

type Resolver struct {
	addr     string
	interval time.Duration
	logger   log.Logger
	canceler run.Canceler
}

func (r *Resolver) Resolve() ([]transport.Address, error) {
	addrs, err := r.resolve(r.addr)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *Resolver) Watch(notify func([]transport.Address)) {
	if r.canceler == nil {
		r.canceler = run.Schedule(r.interval, func() {
			addrs, err := r.Resolve()
			if err != nil {
				r.logger.Error("dns > Failed to refresh addresses: ", err)
			}
			notify(addrs)
		}, nil)
	}
}

func (r *Resolver) resolve(addr string) (addrs []transport.Address, err error) {
	var (
		host, port string
		ips        []string
	)

	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err = net.LookupHost(host)
	if err == nil {
		addrs = make([]transport.Address, len(ips))
		for i, ip := range ips {
			addrs[i] = transport.Address{
				URL: net.JoinHostPort(ip, port),
			}
		}
	}
	return
}

func (r *Resolver) Close() {
	if r.canceler != nil {
		r.canceler.Cancel()
		r.logger.Debug("dns > Resolver stopped")
	}
}

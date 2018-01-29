package etcd

import (
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/rpc/registry"
)

type Registry struct {
	addrs []string
}

type Builder struct {
}

func (Builder) Name() string {
	return "etcd2"
}

func (Builder) Build(name string, opts data.Map) registry.Registry {
	panic("implement me")
}

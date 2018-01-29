package etcd

import (
	"testing"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/net/transport"
)

func TestRegistry(t *testing.T) {
	s := registry.Server{
		Name:    "demo",
		Version: "1.0.0",
		Addresses: []transport.Address{
			{URL: "localhost:9000"},
		},
	}
	opts := data.Map{
		"address": "192.168.50.57:12379",
	}
	r := registry.Get("etcd").Build(s, opts)
	r.Register()
}

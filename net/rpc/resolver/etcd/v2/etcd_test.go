package etcd

import (
	"testing"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/rpc/resolver"
)

func TestResolver(t *testing.T) {
	c := resolver.Client{
		Server:  "demo",
		Version: ">=1.0.0",
	}
	opts := data.Map{
		"address": "192.168.50.57:12379",
	}
	r, err := resolver.Get("etcd").Build(c, opts)
	if err != nil {
		t.Fatal(err)
	}
	addrs, err := r.Resolve()
	t.Log(addrs, err)
}

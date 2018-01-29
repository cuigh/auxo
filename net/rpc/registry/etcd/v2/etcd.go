package etcd

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/etcd/v2"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/net/rpc/registry/valkeyrie"
)

func init() {
	etcd.Register()
	registry.Register(&valkeyrie.Builder{Backend: store.ETCD})
}

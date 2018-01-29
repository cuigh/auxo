package etcd

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/etcd/v3"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/net/rpc/registry/valkeyrie"
)

func init() {
	etcdv3.Register()
	registry.Register(&valkeyrie.Builder{Backend: store.ETCDV3})
}

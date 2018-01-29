package etcd

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/etcd/v3"
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/rpc/resolver/valkeyrie"
)

func init() {
	etcdv3.Register()
	resolver.Register(&valkeyrie.Builder{Backend: store.ETCDV3})
}

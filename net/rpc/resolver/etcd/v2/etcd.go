package etcd

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/etcd/v2"
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/rpc/resolver/valkeyrie"
)

func init() {
	etcd.Register()
	resolver.Register(&valkeyrie.Builder{Backend: store.ETCD})
}

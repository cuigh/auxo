package consul

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/consul"
	"github.com/cuigh/auxo/net/rpc/registry"
	"github.com/cuigh/auxo/net/rpc/registry/valkeyrie"
)

func init() {
	consul.Register()
	registry.Register(&valkeyrie.Builder{Backend: store.CONSUL})
}

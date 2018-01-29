package consul

import (
	"github.com/abronan/valkeyrie/store"
	"github.com/abronan/valkeyrie/store/consul"
	"github.com/cuigh/auxo/net/rpc/resolver"
	"github.com/cuigh/auxo/net/rpc/resolver/valkeyrie"
)

func init() {
	consul.Register()
	resolver.Register(&valkeyrie.Builder{Backend: store.CONSUL})
}

package dns

import (
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/transport"
	"testing"
	"time"
)

func TestResolver_Resolve(t *testing.T) {
	r := newResolver()
	addrs, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range addrs {
		t.Log(addr.URL)
	}
}

func TestResolver_Watch(t *testing.T) {
	r := newResolver()
	r.Watch(func(addrs []transport.Address) {
		for _, addr := range addrs {
			t.Log(addr.URL)
		}
	})
	time.Sleep(10 * time.Second)
}

func newResolver() *Resolver {
	return &Resolver{
		addrs: []transport.Address{
			{URL: "localhost:8888"},
		},
		interval: 3 * time.Second,
		logger:   log.Get(PkgName),
	}
}

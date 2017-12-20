package rpc

import (
	"math/rand"
	"strings"
	"sync/atomic"

	"github.com/cuigh/auxo/data"
)

// todo: move to balancer pkg
var balancers = map[string]BalancerBuilder{}

func init() {
	RegisterBalancer(randomBalancerBuilder{})
	RegisterBalancer(roundRobinBalancerBuilder{})
}

func RegisterBalancer(b BalancerBuilder) {
	balancers[strings.ToLower(b.Name())] = b
}

func GetBalancer(name string) BalancerBuilder {
	return balancers[strings.ToLower(name)]
}

// BalancerBuilder creates a balancer.
type BalancerBuilder interface {
	// Name returns the name of balancers built by this builder.
	// It will be used to pick balancers (for example in service config).
	Name() string
	// Build creates a new balancer with the options.
	Build(opts data.Map) Balancer
}

type BalancerOptions struct {
}

type Balancer interface {
	Update(nodes []*Node)
	Next() (*Node, error)
}

type randomBalancerBuilder struct {
}

func (randomBalancerBuilder) Name() string {
	return "random"
}

func (randomBalancerBuilder) Build(opts data.Map) Balancer {
	return new(randomBalancer)
}

type randomBalancer struct {
	nodes []*Node
}

func (b *randomBalancer) Update(nodes []*Node) {
	b.nodes = nodes
}

func (b *randomBalancer) Next() (*Node, error) {
	nodes := b.nodes
	if l := len(nodes); l > 0 {
		return nodes[rand.Intn(l)], nil
	}
	return nil, ErrNodeUnavailable
}

type roundRobinBalancerBuilder struct {
}

func (roundRobinBalancerBuilder) Name() string {
	return "round_robin"
}

func (roundRobinBalancerBuilder) Build(opts data.Map) Balancer {
	return new(roundRobinBalancer)
}

type roundRobinBalancer struct {
	nodes   []*Node
	counter int64
}

func (b *roundRobinBalancer) Update(nodes []*Node) {
	b.counter = -1
	b.nodes = nodes
}

func (b *roundRobinBalancer) Next() (*Node, error) {
	nodes := b.nodes
	if l := len(nodes); l > 0 {
		i := atomic.AddInt64(&b.counter, 1) % int64(l)
		return nodes[i], nil
	}
	return nil, ErrNodeUnavailable
}

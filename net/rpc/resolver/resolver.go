package resolver

import (
	"strings"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/transport"
)

var (
	builders = make(map[string]Builder)
)

func init() {
	Register(directBuilder{})
}

type Client struct {
	Server  string
	Version string
	Group   string
	Codec   string
}

// Resolver defines interfaces for nodes discovery.
type Resolver interface {
	// Resolve tries to lookup nodes right now.
	Resolve() ([]transport.Address, error)
	// Watch register a watch to Resolver.
	Watch(notify func([]transport.Address))
	// Close stop the watch.
	Close()
}

// Builder creates a Resolver.
type Builder interface {
	// Name returns the name of Resolvers built by this builder.
	Name() string
	// Build creates a new Resolver with the options.
	Build(c Client, opts data.Map) (Resolver, error)
}

// Register registers the Resolver builder to the builder map.
// b.Name (lower-cased) will be used as the name registered with
// this builder.
func Register(b Builder) {
	builders[strings.ToLower(b.Name())] = b
}

// Get returns the resolver builder registered with the given name.
// Note that the compare is done in a case-insenstive fashion.
// If no builder is register with the name, nil will be returned.
func Get(name string) Builder {
	return builders[strings.ToLower(name)]
}

type directBuilder struct{}

func (directBuilder) Name() string {
	return "direct"
}

func (directBuilder) Build(_ Client, opts data.Map) (Resolver, error) {
	addrs := opts.Get("addresses").([]transport.Address)
	return &directResolver{addrs: addrs}, nil
}

type directResolver struct {
	addrs []transport.Address
}

func (r *directResolver) Resolve() ([]transport.Address, error) {
	return r.addrs, nil
}

func (r *directResolver) Watch(notify func([]transport.Address)) {
}

func (r *directResolver) Close() {
}

func Direct(addrs ...transport.Address) Resolver {
	return &directResolver{addrs}
}

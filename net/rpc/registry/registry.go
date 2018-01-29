package registry

import (
	"strings"

	"fmt"

	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/transport"
)

var (
	builders = make(map[string]Builder)
)

// Server holds the server information for registering.
type Server struct {
	Name      string
	Version   string
	Addresses []transport.Address
	// Options returns additional options need send to registry with all addresses.
	Options func() data.Map
}

// Registry defines interfaces for register server.
type Registry interface {
	// Register send all addresses to registry and keep refreshing.
	Register()
	//Heartbeat()
	// Offline remove server addresses from registry and stop refreshing immediately.
	Offline() error
	// Online recover server registering.
	Online() error
	// Close stop the register.
	Close()
}

// Builder creates a registry.
type Builder interface {
	// Name returns the name of registry built by this builder.
	Name() string
	// Build creates a new Registry with the options.
	//Build(name string, opts data.Map) Registry
	Build(server Server, opts data.Map) (Registry, error)
}

// Register registers the registry builder to the registry map.
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

type fakeRegistryBuilder struct {
}

func (fakeRegistryBuilder) Name() string {
	return "fake"
}

func (fakeRegistryBuilder) Build(server Server, _ data.Map) (Registry, error) {
	return &fakeRegistry{s: server}, nil
}

// fakeRegistry is a fake registry for demonstrating how to implement.
type fakeRegistry struct {
	closer data.Chan
	s      Server
}

func (r *fakeRegistry) Register() {
	r.register()
	r.closer = make(data.Chan)
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r.register()
			case <-r.closer:
				return
			}
		}
	}()
}

func (r *fakeRegistry) register() {
	opts := data.Map{
		"version": r.s.Version,
	}
	opts.Merge(r.s.Options())
	fmt.Println("registry > register: "+r.s.Name, "|", opts)
}

func (r *fakeRegistry) Offline() error {
	fmt.Println("registry > offline: " + r.s.Name)
	close(r.closer)
	return nil
}

func (r *fakeRegistry) Online() error {
	fmt.Println("registry > online: " + r.s.Name)
	r.Register()
	return nil
}

func (r *fakeRegistry) Close() {
	fmt.Println("registry > close: " + r.s.Name)
	close(r.closer)
}

func init() {
	Register(fakeRegistryBuilder{})
}

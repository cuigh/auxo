package limiter

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

// This package was inspired by: github.com/ulule/limiter

// Rate is the rate of limiter.
type Rate struct {
	Period time.Duration
	Count  int64
}

// ParseRate try parsing a string to Rate. Valid format: 100/5s, 1000/1m...
func ParseRate(s string) (r *Rate, err error) {
	args := strings.Split(s, "/")
	if len(args) != 2 {
		return nil, errors.New("invalid rate: " + s)
	}

	count, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("invalid period: " + args[0])
	}

	period, err := time.ParseDuration(args[1])
	if err != nil {
		return nil, errors.New("invalid count: " + args[1])
	}

	if period <= 0 || count <= 0 {
		return r, errors.New("invalid rate: " + s)
	}
	return &Rate{Period: period, Count: count}, nil
}

// Store is the common interface for limiter stores.
type Store interface {
	// Get returns the limited resource for given key.
	Get(ctx context.Context, rate *Rate, key string) (Resource, error)
	// Peek returns the limited resource for given key, without modification on current values.
	Peek(ctx context.Context, rate *Rate, key string) (Resource, error)
}

// Resource is the limit resource information.
type Resource struct {
	Limit     int64
	Remaining int64
	// reset time in UnixNano
	Reset int64
}

// OK returns whether the resource is valid.
func (r Resource) OK() bool {
	return r.Remaining >= 0
}

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	Store
	*Rate
}

// New creates an instance of Limiter.
func New(store Store, rate *Rate) *Limiter {
	return &Limiter{
		Store: store,
		Rate:  rate,
	}
}

// Get returns the limited resource for given key.
func (l *Limiter) Get(ctx context.Context, key string) (Resource, error) {
	return l.Store.Get(ctx, l.Rate, key)
}

// Peek returns the limited resource for given key, without modification on current values.
func (l *Limiter) Peek(ctx context.Context, key string) (Resource, error) {
	return l.Store.Peek(ctx, l.Rate, key)
}

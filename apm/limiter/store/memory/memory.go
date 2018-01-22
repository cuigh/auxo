package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuigh/auxo/apm/limiter"
)

type counter struct {
	expiry    int64
	remaining int64
}

func (c *counter) valid(now int64) bool {
	return c.expiry > now
}

func (c *counter) reset(rate *limiter.Rate, now int64) {
	c.expiry = now + rate.Period.Nanoseconds()
	c.remaining = rate.Count
}

type Options struct {
	PruneInterval time.Duration
}

type Store struct {
	opts Options

	locker   sync.RWMutex
	counters map[string]*counter

	pruneFlag int32
	pruneTime int64
}

func New(opts ...Options) *Store {
	s := &Store{
		counters: make(map[string]*counter),
	}
	if len(opts) > 0 {
		s.opts = opts[0]
	}
	if s.opts.PruneInterval <= 0 {
		s.opts.PruneInterval = time.Minute * 10
	}
	s.pruneTime = time.Now().Add(s.opts.PruneInterval).UnixNano()
	return s
}

func (s *Store) Get(ctx context.Context, rate *limiter.Rate, key string) (res limiter.Resource, err error) {
	res.Limit = rate.Count
	now := time.Now().UnixNano()

	s.locker.Lock()

	c := s.counters[key]
	if c == nil {
		c = new(counter)
		c.reset(rate, now)
		s.counters[key] = c
	} else if !c.valid(now) {
		c.reset(rate, now)
	}
	c.remaining--

	res.Reset = c.expiry
	res.Remaining = c.remaining

	s.locker.Unlock()

	// auto clean expired counters
	if s.pruneTime < now && atomic.CompareAndSwapInt32(&s.pruneFlag, 0, 1) {
		go s.prune()
	}
	return
}

func (s *Store) Peek(ctx context.Context, rate *limiter.Rate, key string) (res limiter.Resource, err error) {
	s.locker.RLock()
	c := s.counters[key]
	s.locker.RUnlock()

	res.Limit = rate.Count
	now := time.Now().UnixNano()
	if c == nil || !c.valid(now) {
		res.Remaining = rate.Count
		res.Reset = now + rate.Period.Nanoseconds()
	} else {
		res.Remaining = c.remaining
		res.Reset = c.expiry
	}
	return
}

func (s *Store) prune() {
	s.locker.Lock()
	defer s.locker.Unlock()

	now := time.Now().UnixNano()
	for key, value := range s.counters {
		if !value.valid(now) {
			delete(s.counters, key)
		}
	}
	s.pruneFlag = 0
}

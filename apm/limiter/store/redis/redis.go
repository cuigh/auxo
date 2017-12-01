package redis

import (
	"context"
	"strings"
	"time"

	"github.com/cuigh/auxo/apm/limiter"
	"github.com/cuigh/auxo/db/redis"
)

type Options struct {
	Prefix string
}

func (o *Options) ensure(opts ...Options) {
	if len(opts) > 0 {
		*o = opts[0]
	}

	if o.Prefix == "" {
		o.Prefix = "limiter:"
	} else if !strings.HasSuffix(o.Prefix, ":") {
		o.Prefix = o.Prefix + ":"
	}
}

type Store struct {
	Options
	redis.Client
}

func New(db string, opts ...Options) (*Store, error) {
	c, err := redis.Open(db)
	if err != nil {
		return nil, err
	}

	s := &Store{
		Client: c,
	}
	s.Options.ensure(opts...)
	return s, nil
}

func (s *Store) Get(ctx context.Context, rate *limiter.Rate, key string) (res limiter.Resource, err error) {
	res.Limit = rate.Count
	key = s.Options.Prefix + key
	cmd := s.Client.Eval(`local current
current = redis.call("incr",KEYS[1])
if tonumber(current) == 1 then
    redis.call("pexpire",KEYS[1], ARGV[1])
end
return {current, redis.call("pttl",KEYS[1])}`, []string{key}, rate.Period.Nanoseconds()/1e6)

	var r interface{}
	r, err = cmd.Result()
	if err != nil {
		return
	}

	slice := r.([]interface{})
	d := time.Duration(slice[1].(int64)) * time.Millisecond
	res.Remaining = rate.Count - slice[0].(int64)
	res.Reset = time.Now().Add(d).UnixNano()
	return
	//pipe := s.Client.TxPipeline()
	//count := pipe.Incr(key)
	//expiry := pipe.PTTL(key)
}

func (s *Store) Peek(ctx context.Context, rate *limiter.Rate, key string) (res limiter.Resource, err error) {
	res.Limit = rate.Count
	key = s.Options.Prefix + key

	pipe := s.Client.TxPipeline()
	count := pipe.Get(key)
	expiry := pipe.PTTL(key)

	_, err = pipe.Exec()
	if err != nil && err != redis.Nil {
		return
	}

	var n int64
	n, err = count.Int64()
	if err != nil && err != redis.Nil {
		return
	}
	res.Remaining = rate.Count - n

	var d time.Duration
	d, err = expiry.Result()
	if err != nil && err != redis.Nil {
		return
	}
	if d < 0 {
		d = rate.Period
	}
	res.Reset = time.Now().Add(d).UnixNano()
	return
}

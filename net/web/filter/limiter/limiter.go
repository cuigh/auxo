package limiter

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cuigh/auxo/apm/limiter"
	"github.com/cuigh/auxo/net/web"
)

const (
	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"
)

var ErrorTooManyRequests = web.NewError(http.StatusTooManyRequests)

type Keyer func(ctx web.Context) string

func HandlerName(ctx web.Context) string {
	return ctx.Handler().Name()
}

func IP(ctx web.Context) string {
	return ctx.RealIP()
}

type Option func(*Options)

func Key(k Keyer) Option {
	return func(opts *Options) {
		if k != nil {
			opts.key = k
		}
	}
}

type Options struct {
	key Keyer
}

type SimpleLimiter struct {
	limiter *limiter.Limiter
	opts    Options
}

func NewSimple(store limiter.Store, period time.Duration, count int64, opts ...Option) *SimpleLimiter {
	sl := &SimpleLimiter{
		limiter: limiter.New(store, &limiter.Rate{Period: period, Count: count}),
		opts:    Options{key: IP},
	}
	for _, opt := range opts {
		opt(&sl.opts)
	}
	return sl
}

func (sl *SimpleLimiter) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx web.Context) error {
		res, err := sl.limiter.Get(ctx.Request().Context(), sl.opts.key(ctx))
		if err != nil {
			return err
		}

		ctx.SetHeader(headerRateLimit, strconv.FormatInt(res.Limit, 10))
		ctx.SetHeader(headerRateReset, strconv.FormatInt(res.Reset/1e9, 10)) // Unix seconds
		if !res.OK() {
			ctx.SetHeader(headerRateRemaining, "0")
			return ErrorTooManyRequests
		}
		ctx.SetHeader(headerRateRemaining, strconv.FormatInt(res.Remaining, 10))
		return next(ctx)
	}
}

type AutoLimiter struct {
	store limiter.Store
	opts  Options

	locker   sync.RWMutex
	limiters map[string]*limiter.Limiter
}

func NewAuto(store limiter.Store, opts ...Option) *AutoLimiter {
	if store == nil {
		panic(errors.New("limiter: store must be set"))
	}

	l := &AutoLimiter{
		store:    store,
		limiters: make(map[string]*limiter.Limiter),
		opts:     Options{key: IP},
	}
	for _, opt := range opts {
		opt(&l.opts)
	}
	return l
}

func (al *AutoLimiter) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx web.Context) error {
		tag := ctx.Handler().Option("limiter")
		if tag == "" {
			return next(ctx)
		}

		rl, err := al.getLimiter(ctx, tag)
		if err != nil {
			return err
		}

		res, err := rl.Get(ctx.Request().Context(), al.opts.key(ctx))
		if err != nil {
			return err
		}

		ctx.SetHeader(headerRateLimit, strconv.FormatInt(res.Limit, 10))
		ctx.SetHeader(headerRateReset, strconv.FormatInt(res.Reset/1e9, 10)) // Unix seconds
		if !res.OK() {
			ctx.SetHeader(headerRateRemaining, "0")
			return ErrorTooManyRequests
		}
		ctx.SetHeader(headerRateRemaining, strconv.FormatInt(res.Remaining, 10))
		return next(ctx)
	}
}

func (al *AutoLimiter) getLimiter(ctx web.Context, tag string) (lr *limiter.Limiter, err error) {
	al.locker.RLock()
	lr = al.limiters[ctx.Handler().Name()]
	al.locker.RUnlock()
	if lr != nil {
		return
	}

	al.locker.Lock()
	defer al.locker.Unlock()

	if lr = al.limiters[ctx.Handler().Name()]; lr != nil {
		return
	}

	// tag: 100/1s, 1000/1m
	var rate *limiter.Rate
	rate, err = limiter.ParseRate(tag)
	if err != nil {
		return
	}

	lr = limiter.New(al.store, rate)
	al.limiters[ctx.Handler().Name()] = lr
	return
}

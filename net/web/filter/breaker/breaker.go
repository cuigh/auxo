package breaker

import (
	"strings"
	"sync"

	"github.com/cuigh/auxo/apm/breaker"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web"
	"github.com/cuigh/auxo/util/cast"
)

const PkgName = "auxo.net.web.filter.breaker"

type SimpleBreaker struct {
	Breaker  *breaker.Breaker
	Fallback web.HandlerFunc
}

func (sb *SimpleBreaker) Try(handler web.HandlerFunc, ctx web.Context, logger log.Logger) error {
	if sb.Fallback == nil {
		return sb.Breaker.Try(func() error {
			return handler(ctx)
		})
	} else {
		return sb.Breaker.Try(func() error {
			return handler(ctx)
		}, func(err error) error {
			log.Get(PkgName).Debug("breaker > fallback for error: ", err)
			return sb.Fallback(ctx)
		})
	}
}

func (sb *SimpleBreaker) Apply(next web.HandlerFunc) web.HandlerFunc {
	logger := log.Get(PkgName)
	return func(ctx web.Context) error {
		return sb.Try(next, ctx, logger)
	}
}

type AutoBreaker struct {
	locker   sync.Mutex
	breakers map[string]*SimpleBreaker
}

func NewAuto() *AutoBreaker {
	return &AutoBreaker{
		breakers: make(map[string]*SimpleBreaker),
	}
}

func (ab *AutoBreaker) Apply(next web.HandlerFunc) web.HandlerFunc {
	logger := log.Get(PkgName)
	return func(ctx web.Context) error {
		// `breaker:"c=100,fallback"`
		// `breaker:"f=100,fallback"`
		// `breaker:"r=0.2,fallback"`
		// `breaker:",fallback"`
		// `breaker:"r=0.2"`
		tag := ctx.Handler().Option("breaker")
		if tag == "" {
			return next(ctx)
		}

		sb, err := ab.getBreaker(ctx, tag)
		if err != nil {
			return err
		}
		return sb.Try(next, ctx, logger)
	}
}

func (ab *AutoBreaker) getBreaker(ctx web.Context, tag string) (sb *SimpleBreaker, err error) {
	name := ctx.Handler().Name()
	sb = ab.breakers[name]
	if sb != nil {
		return
	}

	ab.locker.Lock()
	defer ab.locker.Unlock()

	sb = ab.breakers[name]
	if sb != nil {
		return
	}

	var (
		args     = strings.Split(tag, ",")
		cond     breaker.Condition
		fallback string
	)

	switch len(args) {
	case 1:
		cond, err = ab.parseCondition(args[0])
	case 2:
		cond, err = ab.parseCondition(args[0])
		fallback = args[1]
	default:
		return nil, errors.New("invalid breaker tag: " + tag)
	}

	if err != nil {
		return nil, err
	}

	sb = &SimpleBreaker{
		Breaker: breaker.NewBreaker(name, cond, breaker.Options{}),
	}
	if fallback != "" {
		sb.Fallback = ctx.Server().Handler(fallback).Action()
	}
	ab.breakers[name] = sb
	return sb, err
}

func (ab *AutoBreaker) parseCondition(s string) (cond breaker.Condition, err error) {
	if s == "" {
		return breaker.ErrorRate(0.5, 10), nil
	}

	// todo: validate condition
	// c=100, f=100, r=0.2
	option := data.ParseOption(s, "=")
	switch option.Name {
	case "r":
		cond = breaker.ErrorRate(cast.ToFloat32(option.Value), 10)
	case "f":
		cond = breaker.ErrorCount(cast.ToUint32(option.Value))
	case "c":
		cond = breaker.ConsecutiveErrorCount(cast.ToUint32(option.Value))
	}
	return
}

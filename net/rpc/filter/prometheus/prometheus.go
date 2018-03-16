package prometheus

import (
	"strconv"
	"time"

	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/rpc"
	"github.com/prometheus/client_golang/prometheus"
)

const PkgName = "auxo.net.rpc.filter.prometheus"

type Options struct {
	name string
}

type Option func(*Options)

func Name(name string) Option {
	return func(opts *Options) {
		if name != "" {
			opts.name = name
		}
	}
}

func Client(opts ...Option) rpc.CFilter {
	options := &Options{
		name: "rpc_client",
	}
	for _, opt := range opts {
		opt(options)
	}
	reqCounter := registerCounter(options, "requests_total", "How many RPC requests processed, partitioned by error code and action.", "server", "action", "code")
	reqTime := registerSummary(options, "request_duration_seconds", "The RPC request latencies in seconds, partitioned by action.", "server", "action")

	return func(next rpc.CHandler) rpc.CHandler {
		return func(c *rpc.Call) (err error) {
			start := time.Now()
			defer func() {
				status := strconv.Itoa(int(rpc.StatusOf(err)))
				d := float64(time.Since(start)) / float64(time.Second)
				action := texts.Join(".", c.Request().Head.Service, c.Request().Head.Method)
				reqCounter.WithLabelValues(c.Server(), action, status).Inc()
				reqTime.WithLabelValues(c.Server(), action).Observe(d)
			}()

			err = next(c)
			return
		}
	}
}

func Server(opts ...Option) rpc.SFilter {
	options := &Options{
		name: "rpc_server",
	}
	for _, opt := range opts {
		opt(options)
	}
	reqCounter := registerCounter(options, "requests_total", "How many RPC requests processed, partitioned by error code and action.", "code", "action")
	reqTime := registerSummary(options, "request_duration_seconds", "The RPC request latencies in seconds, partitioned by action.", "action")

	return func(next rpc.SHandler) rpc.SHandler {
		return func(c rpc.Context) (r interface{}, err error) {
			start := time.Now()
			defer func() {
				status := strconv.Itoa(int(rpc.StatusOf(err)))
				d := float64(time.Since(start)) / float64(time.Second)
				reqCounter.WithLabelValues(status, c.Action().Name()).Inc()
				reqTime.WithLabelValues(c.Action().Name()).Observe(d)
			}()

			r, err = next(c)
			return
		}
	}
}

func registerCounter(opts *Options, name, help string, labels ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: opts.name,
			Name:      name,
			Help:      help,
		},
		labels,
	)
	registerCollector(name, counter)
	return counter
}

func registerSummary(opts *Options, name, help string, labels ...string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: opts.name,
			Name:      name,
			Help:      help,
		},
		labels,
	)
	registerCollector(name, summary)
	return summary
}

func registerCollector(name string, collector prometheus.Collector) {
	if err := prometheus.Register(collector); err != nil {
		log.Get(PkgName).Errorf("rpc > prometheus: failed to register collector '%v': %v", name, err)
	} else {
		log.Get(PkgName).Infof("rpc > prometheus: collector '%v' registered", name)
	}
}

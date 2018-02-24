package prometheus

import (
	"strconv"
	"time"

	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/net/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const PkgName = "auxo.net.web.filter.prometheus"

type Option func(*Prometheus)

func Name(name string) Option {
	return func(p *Prometheus) {
		if name != "" {
			p.name = name
		}
	}
}

type Prometheus struct {
	name       string
	logger     log.Logger
	reqCounter *prometheus.CounterVec
	reqTime    *prometheus.SummaryVec
	//reqSize    prometheus.Summary
	respSize prometheus.Summary
}

func New(opts ...Option) *Prometheus {
	p := &Prometheus{
		name:   "http",
		logger: log.Get(PkgName),
	}
	for _, opt := range opts {
		opt(p)
	}
	p.registerMetrics()
	return p
}

// Apply implements `web.Filter` interface.
func (p *Prometheus) Apply(next web.HandlerFunc) web.HandlerFunc {
	return func(c web.Context) error {
		start := time.Now()
		defer func() {
			status := strconv.Itoa(c.Response().Status())
			d := float64(time.Since(start)) / float64(time.Second)
			size := float64(c.Response().Size())

			p.reqCounter.WithLabelValues(status, c.Request().Method, c.Handler().Name()).Inc()
			p.reqTime.WithLabelValues(c.Handler().Name()).Observe(d)
			p.respSize.Observe(size)
		}()

		return next(c)
	}
}

func (p *Prometheus) registerMetrics() {
	p.reqCounter = p.registerCounter("requests_total", "How many HTTP requests processed, partitioned by status code and HTTP method.")
	p.reqTime = p.registerSummaryVec("request_duration_seconds", "The HTTP request latencies in seconds, partitioned by handler.", "handler")
	//p.reqSize = p.registerSummary("request_size_bytes", "The HTTP request sizes in bytes.")
	p.respSize = p.registerSummary("response_size_bytes", "The HTTP response sizes in bytes.")
}

func (p *Prometheus) registerCounter(name, help string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: p.name,
			Name:      name,
			Help:      help,
		},
		[]string{"code", "method", "handler"},
	)
	p.registerCollector(name, counter)
	return counter
}

func (p *Prometheus) registerSummary(name, help string) prometheus.Summary {
	summary := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: p.name,
			Name:      name,
			Help:      help,
		},
	)
	p.registerCollector(name, summary)
	return summary
}

func (p *Prometheus) registerSummaryVec(name, help string, labels ...string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: p.name,
			Name:      name,
			Help:      help,
		},
		labels,
	)
	p.registerCollector(name, summary)
	return summary
}

func (p *Prometheus) registerCollector(name string, collector prometheus.Collector) {
	if err := prometheus.Register(collector); err != nil {
		p.logger.Errorf("web > prometheus: failed to register collector '%v': %v", name, err)
	} else {
		p.logger.Infof("web > prometheus: collector '%v' registered", name)
	}
}

//func (p *Prometheus) computeApproximateRequestSize(r *http.Request) <-chan int {
//	// Get URL length in current go routine for avoiding a race condition.
//	// HandlerFunc that runs in parallel may modify the URL.
//	s := 0
//	if r.URL != nil {
//		s += len(r.URL.String())
//	}
//
//	out := make(chan int, 1)
//
//	go func() {
//		s += len(r.Method)
//		s += len(r.Proto)
//		for name, values := range r.Header {
//			s += len(name)
//			for _, value := range values {
//				s += len(value)
//			}
//		}
//		s += len(r.Host)
//
//		// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.
//
//		if r.ContentLength != -1 {
//			s += int(r.ContentLength)
//		}
//		out <- s
//		close(out)
//	}()
//
//	return out
//}

func Handler() web.HandlerFunc {
	h := promhttp.Handler()
	return web.WrapHandler(h)
}

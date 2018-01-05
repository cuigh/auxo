package jaeger

import (
	"io"
	"time"

	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jc "github.com/uber/jaeger-client-go/config"
)

const (
	PkgName   = "auxo.apm.trace.jaeger"
	optionKey = "trace.jaeger"
)

var (
	ErrEmptyName = errors.New("jaeger: empty name")
)

type Options struct {
	Name       string `json:"name" yaml:"name"`
	Enabled    bool   `json:"enabled"yaml:"enabled"`
	RPCMetrics bool   `json:"rpc_metrics" yaml:"rpc_metrics"`
	Sampler    struct {
		Type  string  `json:"type" yaml:"type"`
		Param float64 `json:"param" yaml:"param"`
	} `json:"sampler" yaml:"sampler"`
	Reporter struct {
		Address       string        `json:"address" yaml:"address"`
		FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"`
		Log           bool          `json:"log" yaml:"log"`
		QueueSize     int           `json:"queue_size" yaml:"queue_size"`
	} `json:"reporter" yaml:"reporter"`
}

// fill defaults and validate
func (opts *Options) ensure() error {
	if !opts.Enabled {
		return nil
	}

	if opts.Name == "" {
		opts.Name = config.GetString("name")
	}
	if opts.Name == "" {
		return ErrEmptyName
	}

	//if opts.Reporter.Address == "" {
	//	opts.Reporter.Address = "127.0.0.1"
	//}
	if opts.Reporter.FlushInterval <= 0 {
		opts.Reporter.FlushInterval = 10 * time.Second
	}
	if opts.Reporter.QueueSize <= 0 {
		opts.Reporter.QueueSize = 1000
	}
	return nil
}

// Auto initialize a global tracer and auto-close it on app exit.
func Auto() {
	closer := MustInitGlobal()
	app.OnClose(func() {
		err := closer.Close()
		if err != nil {
			log.Get(PkgName).Warn(err)
		}
	})
}

func InitGlobal(options ...Options) (io.Closer, error) {
	opts, err := loadOptions(options)
	if err != nil {
		return nil, err
	}

	cfg := newConfig(&opts)
	return cfg.InitGlobalTracer(opts.Name, jc.Logger(&loggerAdapter{}))
}

func MustInitGlobal(options ...Options) io.Closer {
	c, err := InitGlobal(options...)
	if err != nil {
		panic(err)
	}
	return c
}

func Init(options ...Options) (opentracing.Tracer, io.Closer, error) {
	opts, err := loadOptions(options)
	if err != nil {
		return nil, nil, err
	}

	cfg := newConfig(&opts)
	// Example metrics factory. Use github.com/uber/jaeger-lib/metrics respectively
	// to bind to real logging and metrics frameworks.
	// metricsFactory := metrics.NullFactory
	// jc.Metrics(metricsFactory)
	return cfg.New(opts.Name, jc.Logger(&loggerAdapter{}))
}

func MustInit(options ...Options) (opentracing.Tracer, io.Closer) {
	t, c, err := Init(options...)
	if err != nil {
		panic(err)
	}
	return t, c
}

func loadOptions(options []Options) (opts Options, err error) {
	if len(options) > 0 {
		opts = options[0]
	} else if config.Exist(optionKey) {
		err = config.UnmarshalOption(optionKey, &opts)
		if err != nil {
			err = errors.Wrap(err, "failed to parse options")
			return
		}
	}

	err = opts.ensure()
	return
}

func newConfig(opts *Options) jc.Configuration {
	cfg := jc.Configuration{
		Disabled:   !opts.Enabled,
		RPCMetrics: opts.RPCMetrics,
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName:   "auxo-trace-id",
			JaegerBaggageHeader:      "auxo-bag",
			TraceBaggageHeaderPrefix: "auxo-bag-",
		},
		Reporter: &jc.ReporterConfig{
			LocalAgentHostPort:  opts.Reporter.Address,
			BufferFlushInterval: opts.Reporter.FlushInterval,
			LogSpans:            opts.Reporter.Log,
			QueueSize:           opts.Reporter.QueueSize,
		},
	}
	if opts.Sampler.Type != "" {
		cfg.Sampler = &jc.SamplerConfig{
			Type:  opts.Sampler.Type,
			Param: opts.Sampler.Param,
		}
	}
	return cfg
}

type loggerAdapter struct {
	*log.Logger
}

func (l *loggerAdapter) getLogger() *log.Logger {
	if l.Logger == nil {
		l.Logger = log.Get(PkgName)
	}
	return l.Logger
}

// Error logs a message at error priority
func (l *loggerAdapter) Error(msg string) {
	l.getLogger().Error(msg)
}

// Infof logs a message at info priority
func (l *loggerAdapter) Infof(msg string, args ...interface{}) {
	l.getLogger().Infof(msg, args...)
}

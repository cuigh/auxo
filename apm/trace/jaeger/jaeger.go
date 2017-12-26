package jaeger

import (
	"io"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jc "github.com/uber/jaeger-client-go/config"
)

const (
	PkgName   = "auxo.apm.trace.jaeger"
	optionKey = "trace.jaeger"
)

type Options struct {
	Enabled    bool `yaml:"enabled"`
	RPCMetrics bool `yaml:"rpc_metrics"`
	Sampler    struct {
		Type  string  `yaml:"type"`
		Param float64 `yaml:"param"`
	} `yaml:"sampler"`
	Reporter struct {
		Address       string        `yaml:"address"`
		FlushInterval time.Duration `yaml:"flush_interval"`
		Log           bool          `yaml:"log"`
		QueueSize     int           `yaml:"queue_size"`
	} `yaml:"reporter"`
}

func (opts *Options) ensure() {
	if !opts.Enabled {
		return
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
}

func InitGlobal(name string, options ...Options) (io.Closer, error) {
	if name == "" {
		name = config.GetString("name")
	}

	// Example metrics factory. Use github.com/uber/jaeger-lib/metrics respectively
	// to bind to real logging and metrics frameworks.
	// metricsFactory := metrics.NullFactory

	cfg := loadConfig(options...)
	return cfg.InitGlobalTracer(
		name,
		jc.Logger(getLogger()),
		//jaegercfg.Metrics(metricsFactory),
	)
}

func MustInitGlobal(name string, options ...Options) io.Closer {
	c, err := InitGlobal(name, options...)
	if err != nil {
		panic(err)
	}
	return c
}

func Init(name string, options ...Options) (opentracing.Tracer, io.Closer, error) {
	if name == "" {
		name = config.GetString("name")
	}
	cfg := loadConfig(options...)
	return cfg.New(name, jc.Logger(getLogger()))
}

func MustInit(name string, options ...Options) (opentracing.Tracer, io.Closer) {
	t, c, err := Init(name, options...)
	if err != nil {
		panic(err)
	}
	return t, c
}

func loadConfig(options ...Options) jc.Configuration {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	} else if config.Exist(optionKey) {
		err := config.UnmarshalOption(optionKey, &opts)
		if err != nil {
			log.Get(PkgName).Error("jaeger > parse options failed: ", err)
		}
	}
	opts.ensure()

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

type loggerAdapter log.Logger

func getLogger() jaeger.Logger {
	logger := log.Get(PkgName)
	return (*loggerAdapter)(logger)
}

// Error logs a message at error priority
func (l *loggerAdapter) Error(msg string) {
	(*log.Logger)(l).Error(msg)
}

// Infof logs a message at info priority
func (l *loggerAdapter) Infof(msg string, args ...interface{}) {
	(*log.Logger)(l).Infof(msg, args...)
}

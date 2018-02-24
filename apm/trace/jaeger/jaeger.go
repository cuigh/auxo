package jaeger

import (
	"io"
	"time"

	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/log"
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
		ServerURL string  `json:"server_url" yaml:"server_url"`
		Type      string  `json:"type" yaml:"type"`
		Param     float64 `json:"param" yaml:"param"`
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

// Auto initialize a global tracer on app start and auto-close it on app exit.
func Auto() {
	app.OnInit(func() error {
		if !config.Exist(optionKey) {
			return errors.New("can't find option: " + optionKey)
		}

		var opts Options
		err := config.UnmarshalOption(optionKey, &opts)
		if err != nil {
			return errors.Wrap(err, "failed to parse options")
		}

		err = opts.ensure()
		if err != nil {
			return errors.Wrap(err, "invalid options")
		}

		//metricsFactory := metrics.NullFactory
		tracer, closer, err := newConfig(&opts).New(opts.Name, jc.Logger(&loggerAdapter{}))
		if err != nil {
			return errors.Wrap(err, "failed to initialize jaeger tracer")
		}
		trace.SetTracer(tracer)

		app.OnClose(func() {
			err := closer.Close()
			if err != nil {
				log.Get(PkgName).Warn(err)
			}
		})
		return nil
	})
}

func Init(opts Options) (io.Closer, error) {
	err := opts.ensure()
	if err != nil {
		return nil, err
	}

	tracer, closer, err := newConfig(&opts).New(opts.Name, jc.Logger(&loggerAdapter{}))
	if err != nil {
		return nil, err
	}
	trace.SetTracer(tracer)
	return closer, nil
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
			SamplingServerURL: opts.Sampler.ServerURL,
			Type:              opts.Sampler.Type,
			Param:             opts.Sampler.Param,
		}
	}
	return cfg
}

type loggerAdapter struct {
	log.Logger
}

func (l *loggerAdapter) getLogger() log.Logger {
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

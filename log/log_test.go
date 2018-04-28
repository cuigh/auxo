package log_test

import (
	slog "log"
	"os"
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/test/assert"
)

func init() {
	config.AddFolder("../config/samples")
}

func TestJSONLayout(t *testing.T) {
	l := log.Get("")
	assert.NotNil(t, l)

	l.Debug("debug")
	l.WithField("foo", "bar").Info("info")
	l.Warn("warn")
	l.Error("error")
}

func TestTextLayout(t *testing.T) {
	l := log.Get("auxo.net.web")
	assert.NotNil(t, l)

	l.Debug("debug")
	l.WithField("foo", "bar").Info("info")
	l.Warn("warn")
	l.Error("error")
}

func TestFileLogger(t *testing.T) {
	l := log.Get("test")
	for i := 0; i < 100; i++ {
		l.Info("github.com/cuigh/auxo/log")
	}
}

func BenchmarkLoggerText(b *testing.B) {
	l := log.Get("benchmark1")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("info")
	}
}

func BenchmarkLoggerJSON(b *testing.B) {
	l := log.Get("benchmark2")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("info")
	}
}

func BenchmarkStdLogger(b *testing.B) {
	f, _ := os.Open(os.DevNull)
	slog.SetOutput(f)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		slog.Println("info")
	}
}

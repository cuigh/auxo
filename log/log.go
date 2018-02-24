package log

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
)

const (
	LevelDebug = Level(0)
	LevelInfo  = Level(1)
	LevelWarn  = Level(2)
	LevelError = Level(3)
	LevelPanic = Level(4)
	LevelFatal = Level(5)
	LevelOff   = Level(6)
)

var (
	levelShortNames = [7]string{"D", "I", "W", "E", "P", "F", "O"}
	levelLongNames  = [7]string{"DEBUG", "INFO", "WARN", "ERROR", "PANIC", "FATAL", "OFF"}
)

var (
	mgr           Manager = &manager{}
	defaultLevel          = LevelDebug
	defaultLayout         = "[{L}]{T}: {M}{N}"
)

type Level int8

func parseLevel(l string) (lvl Level, err error) {
	switch strings.ToLower(l) {
	case "debug":
		lvl = LevelDebug
	case "info":
		lvl = LevelInfo
	case "warn":
		lvl = LevelWarn
	case "error":
		lvl = LevelError
	case "panic":
		lvl = LevelPanic
	case "fatal":
		lvl = LevelFatal
	case "off":
		lvl = LevelOff
	default:
		err = errors.New("invalid level: " + l)
	}
	return
}

func SetManager(m Manager) {
	if m != nil {
		mgr = m
	}
}

func SetDefaultLevel(lvl Level) {
	defaultLevel = lvl
}

func Get(name string) Logger {
	return mgr.Get(name)
}

func Find(name string) Logger {
	return mgr.Find(name)
}

func Configure(opts Options) error {
	return mgr.Configure(opts)
}

type Options struct {
	Loggers []LoggerOptions
	Writers []WriterOptions
}

type LoggerOptions struct {
	Name    string
	Level   string
	Writers []string
}

type WriterOptions struct {
	Name string
	// console/file/[omit]
	Type   string
	Layout string
	// text/json, default: text
	Format  string
	Options data.Map
}

type Manager interface {
	Get(name string) Logger
	Find(name string) Logger
	Configure(opts Options) error
}

type manager struct {
	once    sync.Once
	root    *logger
	loggers []*logger
}

func (m *manager) Get(name string) Logger {
	l := m.Find(name)
	if l == nil {
		l = m.root
	}
	return l
}

func (m *manager) Find(name string) Logger {
	m.once.Do(m.initialize)

	for _, l := range m.loggers {
		if strings.HasPrefix(name, l.name) {
			return l
		}
	}
	return nil
}

func (m *manager) Configure(opts Options) error {
	// create writers first
	writers := map[string]*Writer{}
	for _, w := range opts.Writers {
		writer, err := newWriter(w.Name, w.Type, w.Format, w.Layout, w.Options)
		if err != nil {
			return err
		}

		writers[w.Name] = writer
	}

	var (
		root    *logger
		loggers = make([]*logger, len(opts.Loggers))
	)
	for i, li := range opts.Loggers {
		lvl, err := parseLevel(li.Level)
		if err != nil {
			return err
		}

		loggers[i] = &logger{
			name: li.Name,
			lvl:  lvl,
		}

		for _, n := range li.Writers {
			w := writers[n]
			if w == nil {
				return errors.New("writer not found: " + n)
			}
			loggers[i].writers = append(loggers[i].writers, writers[n])
		}

		if li.Name == "" {
			root = loggers[i]
		}
	}

	sort.Slice(loggers, func(i, j int) bool {
		return loggers[i].name > loggers[j].name
	})

	if root == nil {
		fmt.Println("Warn: root logger is not configured, set it to default console logger")
	} else {
		m.root = root
	}
	m.loggers = loggers
	return nil
}

func (m *manager) initialize() {
	if m.loggers == nil && config.Exist("log") {
		opts := Options{}
		err := config.UnmarshalOption("log", &opts)
		if err == nil {
			err = Configure(opts)
		}
		if err != nil {
			fmt.Println("log > Auto configure failed:", err)
		}
	}

	if m.root == nil {
		m.root = m.createDefaultLogger()
	}
}

func (m *manager) createDefaultLogger() *logger {
	w, err := newWriter("console", "console", "text", defaultLayout, data.Map{})
	if err != nil {
		panic(err)
	}
	return &logger{
		lvl:     defaultLevel,
		writers: []*Writer{w},
	}
}

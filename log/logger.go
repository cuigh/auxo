package log

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
)

const defaultLayout = "[{L}]{T}: {M}{N}"

var (
	once            sync.Once
	root            *Logger
	loggers         []*Logger
	levelShortNames = [7]string{"D", "I", "W", "E", "P", "F", "O"}
	levelLongNames  = [7]string{"DEBUG", "INFO", "WARN", "ERROR", "PANIC", "FATAL", "OFF"}
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

type Row struct {
	lvl    Level
	time   time.Time
	msg    string
	fields map[string]interface{}
}

type Logger struct {
	locker sync.Mutex
	name   string
	lvl    Level
	//prefix  string
	writers []*Writer
	row     Row
}

func (l *Logger) Name() string {
	return l.name
}

func (l *Logger) SetName(name string) {
	l.name = name
}

func (l *Logger) Level() Level {
	return l.lvl
}

func (l *Logger) SetLevel(lvl Level) {
	l.lvl = lvl
}

//func (l *Logger) Prefix() string {
//	return l.prefix
//}
//
//func (l *Logger) SetPrefix(prefix string) {
//	l.prefix = prefix
//}

func (l *Logger) IsDebugEnabled() bool {
	return l.lvl <= LevelDebug
}

func (l *Logger) IsInfoEnabled() bool {
	return l.lvl <= LevelInfo
}

func (l *Logger) IsWarnEnabled() bool {
	return l.lvl <= LevelWarn
}

func (l *Logger) Debug(args ...interface{}) {
	l.write(LevelDebug, fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.write(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...interface{}) {
	l.write(LevelInfo, fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.write(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...interface{}) {
	l.write(LevelWarn, fmt.Sprint(args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.write(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.write(LevelError, fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.write(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.write(LevelPanic, s)
	panic(s)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	l.write(LevelPanic, s)
	panic(s)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.write(LevelFatal, fmt.Sprint(args...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.write(LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Write implement io.Writer interface
func (l *Logger) Write(p []byte) (n int, err error) {
	for _, w := range l.writers {
		n, err = w.Output().Write(p)
		if err != nil {
			return
		}
	}
	return
}

func (l *Logger) write(lvl Level, msg string) {
	if l.lvl > lvl {
		return
	}

	l.locker.Lock()
	l.row.lvl = lvl
	l.row.time = time.Now()
	l.row.msg = msg
	for _, w := range l.writers {
		w.Write(&l.row)
	}
	l.locker.Unlock()
}

func Get(name string) (l *Logger) {
	l = Find(name)
	if l == nil {
		l = root
	}
	return
}

func Find(name string) *Logger {
	once.Do(initialize)

	for _, l := range loggers {
		if strings.HasPrefix(name, l.name) {
			return l
		}
	}
	return nil
}

func initialize() {
	if loggers == nil {
		opts := &Config{}
		err := config.UnmarshalOption("log", opts)
		if err == nil {
			err = Configure(opts)
		}
		if err != nil {
			fmt.Println("log > Auto configure failed:", err)
		}
	}

	if root == nil {
		root = createDefaultLogger()
	}
}

func createDefaultLogger() *Logger {
	w, err := newWriter("console", "console", "text", defaultLayout, data.Map{})
	if err != nil {
		panic(err)
	}
	return &Logger{
		lvl:     LevelDebug,
		writers: []*Writer{w},
	}
}

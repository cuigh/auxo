package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Entry interface {
	WithField(key string, value interface{}) Entry
	WithFields(fields map[string]interface{}) Entry
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type Logger interface {
	io.Writer
	Name() string
	Level() Level
	SetLevel(lvl Level)
	IsEnabled(lvl Level) bool
	WithField(key string, value interface{}) Entry
	WithFields(fields map[string]interface{}) Entry
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type logger struct {
	locker sync.Mutex
	name   string
	lvl    Level
	//prefix  string
	writers []*Writer
	entries sync.Pool
}

func (l *logger) Name() string {
	return l.name
}

func (l *logger) Level() Level {
	return l.lvl
}

func (l *logger) SetLevel(lvl Level) {
	l.lvl = lvl
}

//func (l *logger) Prefix() string {
//	return l.prefix
//}
//
//func (l *logger) SetPrefix(prefix string) {
//	l.prefix = prefix
//}

func (l *logger) IsEnabled(lvl Level) bool {
	return l.lvl <= lvl
}

func (l *logger) WithField(key string, value interface{}) Entry {
	e := l.getEntry()
	return e.WithField(key, value)
}

func (l *logger) WithFields(fields map[string]interface{}) Entry {
	e := l.getEntry()
	return e.WithFields(fields)
}

func (l *logger) Debug(args ...interface{}) {
	if l.lvl <= LevelDebug {
		e := l.getEntry()
		e.Debug(args...)
		l.putEntry(e)
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if l.lvl <= LevelDebug {
		e := l.getEntry()
		e.Debugf(format, args...)
		l.putEntry(e)
	}
}

func (l *logger) Info(args ...interface{}) {
	if l.lvl <= LevelInfo {
		e := l.getEntry()
		e.Info(args...)
		l.putEntry(e)
	}
}

func (l *logger) Infof(format string, args ...interface{}) {
	if l.lvl <= LevelInfo {
		e := l.getEntry()
		e.Infof(format, args...)
		l.putEntry(e)
	}
}

func (l *logger) Warn(args ...interface{}) {
	if l.lvl <= LevelWarn {
		e := l.getEntry()
		e.Warn(args...)
		l.putEntry(e)
	}
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if l.lvl <= LevelWarn {
		e := l.getEntry()
		e.Warnf(format, args...)
		l.putEntry(e)
	}
}

func (l *logger) Error(args ...interface{}) {
	if l.lvl <= LevelError {
		e := l.getEntry()
		e.Error(args...)
		l.putEntry(e)
	}
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if l.lvl <= LevelError {
		e := l.getEntry()
		e.Errorf(format, args...)
		l.putEntry(e)
	}
}

func (l *logger) Panic(args ...interface{}) {
	if l.lvl <= LevelPanic {
		e := l.getEntry()
		e.Panic(args...)
		l.putEntry(e)
	}
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if l.lvl <= LevelPanic {
		e := l.getEntry()
		e.Panicf(format, args...)
		l.putEntry(e)
	}
}

func (l *logger) Fatal(args ...interface{}) {
	e := l.getEntry()
	e.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	e := l.getEntry()
	e.Fatalf(format, args...)
}

// Write implement io.Writer interface
func (l *logger) Write(p []byte) (n int, err error) {
	l.locker.Lock()
	for _, w := range l.writers {
		n, err = w.Output().Write(p)
		if err != nil {
			return
		}
	}
	l.locker.Unlock()
	return
}

func (l *logger) getEntry() *entry {
	if e := l.entries.Get(); e != nil {
		return e.(*entry)
	}
	return &entry{logger: l, buf: &bytes.Buffer{}}
}

func (l *logger) putEntry(e *entry) {
	e.fields = nil
	l.entries.Put(e)
}

type entry struct {
	*logger
	locker sync.Mutex
	buf    *bytes.Buffer
	lvl    Level
	time   time.Time
	msg    string
	fields map[string]interface{}
}

func (e *entry) WithField(key string, value interface{}) Entry {
	if e.fields == nil {
		e.fields = map[string]interface{}{key: value}
	} else {
		e.fields[key] = value
	}
	return e
}

func (e *entry) WithFields(fields map[string]interface{}) Entry {
	if e.fields == nil {
		// todo: copy data?
		e.fields = fields
	} else {
		for k, v := range fields {
			e.fields[k] = v
		}
	}
	return e
}

func (e *entry) Debug(args ...interface{}) {
	if e.logger.lvl <= LevelDebug {
		e.write(LevelDebug, fmt.Sprint(args...))
	}
}

func (e *entry) Debugf(format string, args ...interface{}) {
	if e.logger.lvl <= LevelDebug {
		e.write(LevelDebug, fmt.Sprintf(format, args...))
	}
}

func (e *entry) Info(args ...interface{}) {
	if e.logger.lvl <= LevelInfo {
		e.write(LevelInfo, fmt.Sprint(args...))
	}
}

func (e *entry) Infof(format string, args ...interface{}) {
	if e.logger.lvl <= LevelInfo {
		e.write(LevelInfo, fmt.Sprintf(format, args...))
	}
}

func (e *entry) Warn(args ...interface{}) {
	if e.logger.lvl <= LevelWarn {
		e.write(LevelWarn, fmt.Sprint(args...))
	}
}

func (e *entry) Warnf(format string, args ...interface{}) {
	if e.logger.lvl <= LevelWarn {
		e.write(LevelWarn, fmt.Sprintf(format, args...))
	}
}

func (e *entry) Error(args ...interface{}) {
	if e.logger.lvl <= LevelError {
		e.write(LevelError, fmt.Sprint(args...))
	}
}

func (e *entry) Errorf(format string, args ...interface{}) {
	if e.logger.lvl <= LevelError {
		e.write(LevelError, fmt.Sprintf(format, args...))
	}
}

func (e *entry) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	if e.logger.lvl <= LevelPanic {
		e.write(LevelPanic, s)
	}
	panic(s)
}

func (e *entry) Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if e.logger.lvl <= LevelPanic {
		e.write(LevelPanic, s)
	}
	panic(s)
}

func (e *entry) Fatal(args ...interface{}) {
	if e.logger.lvl <= LevelFatal {
		e.write(LevelPanic, fmt.Sprint(args...))
	}
	os.Exit(1)
}

func (e *entry) Fatalf(format string, args ...interface{}) {
	if e.logger.lvl <= LevelFatal {
		e.write(LevelFatal, fmt.Sprintf(format, args...))
	}
	os.Exit(1)
}

func (e *entry) write(lvl Level, msg string) {
	e.locker.Lock()
	e.lvl = lvl
	e.time = time.Now()
	e.msg = msg
	for _, w := range e.writers {
		w.Write(e)
	}
	e.locker.Unlock()
}

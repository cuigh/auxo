package log

import (
	"errors"
	"sort"

	"fmt"

	"github.com/cuigh/auxo/data"
)

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

type Config struct {
	Loggers []LoggerOptions
	Writers []WriterOptions
}

func Configure(cfg *Config) error {
	// create writers first
	writers := map[string]*Writer{}
	for _, w := range cfg.Writers {
		writer, err := newWriter(w.Name, w.Type, w.Format, w.Layout, w.Options)
		if err != nil {
			return err
		}

		writers[w.Name] = writer
	}

	loggers = make([]*Logger, len(cfg.Loggers))
	for i, li := range cfg.Loggers {
		lvl, err := parseLevel(li.Level)
		if err != nil {
			return err
		}

		loggers[i] = &Logger{
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
		fmt.Println("Warn: root logger is not configured!")
	}
	return nil
}

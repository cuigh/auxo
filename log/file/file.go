package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/byte/size"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/ext/times"
	"github.com/cuigh/auxo/util/cast"
)

type writer struct {
	dir               string
	name              string
	ext               string
	rollingSize       int
	rollingTime       time.Duration
	rollingTimeLayout string

	locker  sync.Mutex // protect following
	file    *os.File
	size    int
	index   int
	current time.Time
	next    time.Time
}

func New(options data.Map) (io.Writer, error) {
	name := cast.ToString(options.Get("name"))
	if name == "" {
		return nil, errors.New("missing required option: name")
	}

	w := &writer{}

	rollingSize := cast.ToString(options.Get("rolling.size"))
	if rollingSize != "" {
		s, err := size.Parse(rollingSize)
		if err != nil {
			return nil, errors.New("rolling.size option is invalid: " + rollingSize)
		}
		w.rollingSize = int(s)
	}

	w.rollingTime = cast.ToDuration(options.Get("rolling.time"))
	if w.rollingTime > 0 {
		switch {
		case w.rollingTime >= times.Day:
			w.rollingTimeLayout = "yyyyMMdd"
		case w.rollingTime >= time.Hour:
			w.rollingTimeLayout = "yyyyMMddHH"
		default:
			w.rollingTimeLayout = "yyyyMMddHHmm"
		}
	}

	var base string
	if filepath.IsAbs(name) {
		w.dir = filepath.Dir(name)
		base = filepath.Base(name)
	} else {
		// todo: $HOME/logs/$app as default?
		if logPath, appName := config.GetString("global.log.path"), config.GetString("name"); logPath != "" && appName != "" {
			w.dir = filepath.Join(logPath, appName)
		} else {
			w.dir, _ = os.Getwd()
		}
		base = name
	}
	w.ext = filepath.Ext(base)
	w.name = strings.TrimSuffix(base, w.ext)
	return w, nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.file == nil {
		err = w.openFile()
		if err != nil {
			fmt.Println("open log file failed:", err)
			return
		}
	}

	if err = w.tryRollingTime(); err != nil {
		return
	}
	n, err = w.file.Write(p)
	if n > 0 {
		w.size += n
		if err = w.tryRollingSize(); err != nil {
			return
		}
	}
	return
}

func (w *writer) openFile() error {
	name := w.getFileName(false)
	f, err := files.Open(name)
	if err != nil {
		return err
	}

	fi, err := os.Stat(name)
	if err != nil {
		return err
	}

	w.file, w.size = f, int(fi.Size())
	w.index = 0
	if w.rollingTime > 0 {
		w.current = time.Now().Truncate(w.rollingTime)
		w.next = w.current.Add(w.rollingTime)
	}
	return nil
}

func (w *writer) getFileName(bak bool) string {
	if bak {
		var suffix string
		if w.rollingTime > 0 {
			suffix = "." + times.Format(w.current, w.rollingTimeLayout)
		}
		if w.index > 0 {
			suffix += fmt.Sprintf(".%v", w.index)
		}
		if suffix != "" {
			return filepath.Join(w.dir, w.name+suffix+w.ext)
		}
	}
	return filepath.Join(w.dir, w.name+w.ext)
}

func (w *writer) tryRollingTime() error {
	now := time.Now()
	if w.rollingTime == 0 || w.next.After(now) {
		return nil
	}

	return w.rolling()
}

func (w *writer) tryRollingSize() error {
	if w.rollingSize == 0 || w.size < w.rollingSize {
		return nil
	}

	return w.rolling()
}

func (w *writer) rolling() error {
	path := w.getFileName(true)
	for files.Exist(path) {
		w.index++
		path = w.getFileName(true)
	}

	name := w.file.Name()
	err := w.file.Close()
	if err != nil {
		return err
	}
	w.file = nil

	err = os.Rename(name, path)
	if err != nil {
		return err
	}

	return w.openFile()
}

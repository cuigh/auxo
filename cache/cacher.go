package cache

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/ext/times"
	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/util/cast"
)

// ErrorHandling defines how cache actions behave if the parse fails.
type ErrorHandling int

// These constants cause cache actions to behave as described if the parse fails.
const (
	ErrorSilent ErrorHandling = iota // Do nothing.
	ErrorLog                         // Write log with a descriptive error.
	ErrorPanic                       // Call panic with a descriptive error.
)

// Unmarshal implements config.Unmarshaler interface.
func (e *ErrorHandling) Unmarshal(i interface{}) (err error) {
	if value, ok := i.(string); ok {
		switch value {
		case "silent":
			*e = ErrorSilent
			return
		case "log":
			*e = ErrorLog
			return
		case "panic":
			*e = ErrorPanic
			return
		}
	}
	return fmt.Errorf("unable to unmarshal %#v to ErrorHandling", i)
}

type Keyer func(key string, args ...interface{}) string

type Options struct {
	Name     string
	Provider string
	Enabled  bool
	Error    ErrorHandling
	Prefix   string
	Time     time.Duration
	Options  data.Map
	Keys     map[string]string
}

type Cacher interface {
	// Get returns cached value, the result is assured of not nil.
	Get(key string, args ...interface{}) data.Value
	Set(value interface{}, key string, args ...interface{})
	Exist(key string, args ...interface{}) bool
	Remove(key string, args ...interface{})
	RemoveGroup(key string)
}

type KeyInfo struct {
	Time  time.Duration
	Group string
}

type cacher struct {
	p       Provider
	enabled bool
	eh      ErrorHandling
	keyer   Keyer
	def     *KeyInfo
	keys    map[string]*KeyInfo
	logger  log.Logger
}

func newCacher(p Provider, opts *Options) *cacher {
	c := &cacher{
		p:       p,
		enabled: opts.Enabled,
		eh:      opts.Error,
		keyer:   prefix(opts.Prefix),
		def:     &KeyInfo{Time: opts.Time},
		keys:    make(map[string]*KeyInfo),
		logger:  log.Get(PkgName),
	}
	for key, value := range opts.Keys {
		args := strings.Split(value, ",")
		info := &KeyInfo{
			Time: cast.ToDuration(strings.TrimSpace(args[0]), opts.Time),
		}
		if len(args) > 1 {
			info.Group = strings.TrimSpace(args[1])
		}
		c.keys[key] = info
	}
	return c
}

func (c *cacher) Get(key string, args ...interface{}) (v data.Value) {
	if !c.enabled {
		return data.Nil
	}

	info := c.getInfo(key)
	if info == nil {
		return nil
	}

	var err error
	if info.Group == "" {
		k := c.keyer(key, args...)
		v, err = c.p.Get(k)
		if err != nil {
			c.handleError(err)
		}
	} else {
		g := c.getGroup(info.Group, false)
		if g != "" {
			k := c.appendGroup(c.keyer(key, args...), g)
			v, err = c.p.Get(k)
			if err != nil {
				c.handleError(err)
			}
		}
	}
	return
}

func (c *cacher) Set(value interface{}, key string, args ...interface{}) {
	if !c.enabled {
		return
	}

	info := c.getInfo(key)
	if info == nil {
		return
	}

	var err error
	if info.Group == "" {
		k := c.keyer(key, args...)
		err = c.p.Set(k, value, info.Time)
	} else {
		g := c.getGroup(info.Group, true)
		if g == "" {
			return
		}
		k := c.appendGroup(c.keyer(key, args...), g)
		err = c.p.Set(k, value, info.Time)
	}
	if err != nil {
		c.handleError(err)
	}
}

func (c *cacher) Exist(key string, args ...interface{}) (b bool) {
	if !c.enabled {
		return false
	}

	info := c.getInfo(key)
	if info == nil {
		return false
	}

	var err error
	if info.Group == "" {
		k := c.keyer(key, args...)
		b, err = c.p.Exist(k)
	} else {
		g := c.getGroup(info.Group, false)
		if g != "" {
			k := c.appendGroup(c.keyer(key, args...), g)
			b, err = c.p.Exist(k)
		}
	}
	if err != nil {
		c.handleError(err)
	}
	return
}

func (c *cacher) Remove(key string, args ...interface{}) {
	if !c.enabled {
		return
	}

	info := c.getInfo(key)
	if info == nil {
		return
	}

	var err error
	if info.Group == "" {
		k := c.keyer(key, args...)
		err = c.p.Remove(k)
		if err != nil {
			c.handleError(err)
		}
	} else {
		g := c.getGroup(info.Group, false)
		if g != "" {
			k := c.appendGroup(c.keyer(key, args...), g)
			err = c.p.Remove(k)
			if err != nil {
				c.handleError(err)
			}
		}
	}
}

func (c *cacher) RemoveGroup(key string) {
	if !c.enabled {
		return
	}

	k := c.keyer(key)
	err := c.p.Remove(k)
	if err != nil {
		c.handleError(err)
	}
}

func (c *cacher) getGroup(key string, set bool) (g string) {
	k := c.keyer(key)
	value, err := c.p.Get(k)
	if err == nil {
		if value == nil || value.IsNil() {
			if set {
				g = strconv.FormatInt(time.Now().Unix(), 36)
				// every cache item has it's own cache time, so set a long cache time for version here.
				err = c.p.Set(k, g, times.Days(30))
				if err != nil {
					g = ""
				}
			}
		} else {
			g, err = value.String()
		}
	}
	if err != nil {
		c.handleError(err)
	}
	return
}

func (c *cacher) appendGroup(k, g string) string {
	return k + "@" + g
}

func (c *cacher) getInfo(key string) *KeyInfo {
	info := c.keys[key]
	if info == nil {
		info = c.def
	}
	return info
}

func (c *cacher) handleError(err error) {
	switch c.eh {
	case ErrorLog:
		c.logger.Error("cache > ", err)
	case ErrorPanic:
		panic(err)
	}
}

func prefix(prefix string) Keyer {
	if prefix == "" {
		prefix = "auxo:"
	} else if !strings.HasSuffix(prefix, ":") {
		prefix = prefix + ":"
	}

	return func(key string, args ...interface{}) string {
		length := len(args)
		if length == 0 {
			return prefix + key
		}

		arr := make([]string, length+1)
		arr[0] = key
		for i := 0; i < length; i++ {
			arr[i+1] = cast.ToString(args[i])
		}
		return prefix + strings.Join(arr, "-")
	}
}

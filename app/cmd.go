package app

import (
	"strings"
	"time"

	"github.com/cuigh/auxo/app/flag"
	"github.com/cuigh/auxo/util/cast"
)

var (
	Name     string
	Desc     string
	Action   func(ctx *Context)
	flags    = flag.Default
	children = CommandSet{}
)

func Register(f flag.CommonFlag) {
	flags.Register(f)
}

func AddCommand(sub *Command) {
	children.Add(sub)
}

func StringFlag(full, short, value, usage string) *string {
	return flags.String(full, short, value, usage)
}

func IntFlag(full, short string, value int, usage string) *int {
	return flags.Int(full, short, value, usage)
}

func Int64Flag(full, short string, value int64, usage string) *int64 {
	return flags.Int64(full, short, value, usage)
}

func UintFlag(full, short string, value uint, usage string) *uint {
	return flags.Uint(full, short, value, usage)
}

func Uint64Flag(full, short string, value uint64, usage string) *uint64 {
	return flags.Uint64(full, short, value, usage)
}

func Float64Flag(full, short string, value float64, usage string) *float64 {
	return flags.Float64(full, short, value, usage)
}

func BoolFlag(full, short string, value bool, usage string) *bool {
	return flags.Bool(full, short, value, usage)
}

func DurationFlag(full, short string, value time.Duration, usage string) *time.Duration {
	return flags.Duration(full, short, value, usage)
}

type CommandSet map[string]*Command

func (cs CommandSet) Add(cmd *Command) {
	cs[cmd.Name] = cmd
}

type Command struct {
	Name     string
	Desc     string
	Action   func(ctx *Context)
	flags    *flag.Set
	children CommandSet
}

func NewCommand(name, desc string, action func(ctx *Context)) *Command {
	return &Command{
		Name:     name,
		Desc:     desc,
		Action:   action,
		children: CommandSet{},
		flags:    flag.NewSet(name, desc, flag.ExitOnError),
	}
}

func (c *Command) AddCommand(sub *Command) {
	if c.children == nil {
		c.children = CommandSet{}
	}
	c.children.Add(sub)
}

func (c *Command) Register(f flag.CommonFlag) {
	c.flags.Register(f)
}

func (c *Command) StringFlag(full, short, value, usage string) *string {
	return c.flags.String(full, short, value, usage)
}

func (c *Command) IntFlag(full, short string, value int, usage string) *int {
	return c.flags.Int(full, short, value, usage)
}

func (c *Command) Int64Flag(full, short string, value int64, usage string) *int64 {
	return c.flags.Int64(full, short, value, usage)
}

func (c *Command) UintFlag(full, short string, value uint, usage string) *uint {
	return c.flags.Uint(full, short, value, usage)
}

func (c *Command) Uint64Flag(full, short string, value uint64, usage string) *uint64 {
	return c.flags.Uint64(full, short, value, usage)
}

func (c *Command) Float64Flag(full, short string, value float64, usage string) *float64 {
	return c.flags.Float64(full, short, value, usage)
}

func (c *Command) BoolFlag(full, short string, value bool, usage string) *bool {
	return c.flags.Bool(full, short, value, usage)
}

func (c *Command) DurationFlag(full, short string, value time.Duration, usage string) *time.Duration {
	return c.flags.Duration(full, short, value, usage)
}

type Context struct {
	cmd *Command
}

// Args returns the non-flag arguments.
func (c *Context) Args() []string {
	return c.cmd.flags.Args()
}

// Usage prints usage to os.Stdout.
func (c *Context) Usage() {
	c.cmd.flags.Usage()
}

// Config returns the help argument.
func (c *Context) Help() bool {
	f := c.cmd.flags.Lookup("help")
	if f != nil {
		return cast.ToBool(f.Value.String())
	}
	return false
}

// Config returns the version argument.
func (c *Context) Version() bool {
	f := c.cmd.flags.Lookup("version")
	if f != nil {
		return cast.ToBool(f.Value.String())
	}
	return false
}

// Config returns the config argument.
func (c *Context) Config() string {
	f := c.cmd.flags.Lookup("config")
	if f != nil {
		return f.Value.String()
	}
	return ""
}

// Profiles returns active profiles by command line.
func (c *Context) Profiles() []string {
	f := c.cmd.flags.Lookup("profile")
	if f != nil {
		if s := f.Value.String(); s != "" {
			return strings.Split(s, ",")
		}
	}
	return nil
}

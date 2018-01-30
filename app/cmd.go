package app

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cuigh/auxo/app/flag"
	"github.com/cuigh/auxo/util/cast"
)

var (
	Name     string
	Desc     string
	Action   func(ctx *Context)
	Flags    = flag.Default
	children = CommandSet{}
)

func AddCommand(sub *Command) {
	children.Add(sub)
}

type Command struct {
	Name     string
	Desc     string
	Action   func(ctx *Context)
	Flags    *flag.Set
	children CommandSet
}

func NewCommand(name, desc string, action func(ctx *Context)) *Command {
	return &Command{
		Name:     name,
		Desc:     desc,
		Action:   action,
		children: CommandSet{},
		Flags:    flag.NewSet(name, desc, flag.ExitOnError),
	}
}

func (c *Command) AddCommand(sub *Command) {
	if c.children == nil {
		c.children = CommandSet{}
	}
	c.children.Add(sub)
}

type CommandSet map[string]*Command

func (cs CommandSet) Add(cmd *Command) {
	cs[cmd.Name] = cmd
}

type Context struct {
	cmd *Command
}

// Args returns the non-flag arguments.
func (c *Context) Args() []string {
	return c.cmd.Flags.Args()
}

// Usage prints usage to os.Stdout.
func (c *Context) Usage() {
	c.cmd.Flags.Usage()
	if len(c.cmd.children) > 0 {
		fmt.Print("\nCommands:\n\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for name, child := range c.cmd.children {
			fmt.Fprintln(w, "  "+name+"\t"+child.Desc)
		}
		w.Flush()
	}
}

// Config returns the help argument.
func (c *Context) Help() bool {
	f := c.cmd.Flags.Lookup("help")
	if f != nil {
		return cast.ToBool(f.Value.String())
	}
	return false
}

// Config returns the version argument.
func (c *Context) Version() bool {
	f := c.cmd.Flags.Lookup("version")
	if f != nil {
		return cast.ToBool(f.Value.String())
	}
	return false
}

// Config returns the config argument.
func (c *Context) Config() string {
	f := c.cmd.Flags.Lookup("config")
	if f != nil {
		return f.Value.String()
	}
	return ""
}

// Profiles returns active profiles by command line.
func (c *Context) Profiles() []string {
	f := c.cmd.Flags.Lookup("profile")
	if f != nil {
		if s := f.Value.String(); s != "" {
			return strings.Split(s, ",")
		}
	}
	return nil
}

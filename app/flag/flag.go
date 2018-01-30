// Package implements command-line flag parsing. It's a wrapper of flag pkg in stdlib.
package flag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

type ErrorHandling = flag.ErrorHandling

const (
	ContinueOnError = flag.ContinueOnError // Return a descriptive error.
	ExitOnError     = flag.ExitOnError     // Call os.Exit(2).
	PanicOnError    = flag.PanicOnError    // Call panic with a descriptive error.
)

// CommonFlag represents pre-defined flags.
type CommonFlag int

// Pre-defined common flags
const (
	Help CommonFlag = 1 << iota
	Version
	Profile
	Config
	All = Help | Version | Profile | Config
)

// UsageStyle defines the usage style.
type UsageStyle int32

const (
	// Compact: left-right style
	Compact UsageStyle = iota
	// Manual: up-down style
	Manual
)

func (s UsageStyle) Print(w io.Writer, title, desc string) {
	fmt.Fprint(w, "  "+title)
	switch s {
	case Compact:
		fmt.Fprintln(w, "\t"+desc)
	default:
		fmt.Fprintln(w, "\n      "+desc)
	}
}

var Default = Wrap(flag.CommandLine, "")

// Register adds common flags like help/version/profile/config etc.
func Register(f CommonFlag) {
	Default.Register(f)
}

func Bool(full, short string, value bool, usage string) *bool {
	return Default.Bool(full, short, value, usage)
}

func BoolVar(p *bool, full, short string, value bool, usage string) {
	Default.BoolVar(p, full, short, value, usage)
}

func String(full, short, value, usage string) *string {
	return Default.String(full, short, value, usage)
}

func StringVar(p *string, full, short, value, usage string) {
	Default.StringVar(p, full, short, value, usage)
}

func Int(full, short string, value int, usage string) *int {
	return Default.Int(full, short, value, usage)
}

func IntVar(p *int, full, short string, value int, usage string) {
	Default.IntVar(p, full, short, value, usage)
}

func Int64(full, short string, value int64, usage string) *int64 {
	return Default.Int64(full, short, value, usage)
}

func Int64Var(p *int64, full, short string, value int64, usage string) {
	Default.Int64Var(p, full, short, value, usage)
}

func Uint(full, short string, value uint, usage string) *uint {
	return Default.Uint(full, short, value, usage)
}

func UintVar(p *uint, full, short string, value uint, usage string) {
	Default.UintVar(p, full, short, value, usage)
}

func Uint64(full, short string, value uint64, usage string) *uint64 {
	return Default.Uint64(full, short, value, usage)
}

func Uint64Var(p *uint64, full, short string, value uint64, usage string) {
	Default.Uint64Var(p, full, short, value, usage)
}

func Float64(full, short string, value float64, usage string) *float64 {
	return Default.Float64(full, short, value, usage)
}

func Float64Var(p *float64, full, short string, value float64, usage string) {
	Default.Float64Var(p, full, short, value, usage)
}

func Duration(full, short string, value time.Duration, usage string) *time.Duration {
	return Default.Duration(full, short, value, usage)
}

func DurationVar(p *time.Duration, full, short string, value time.Duration, usage string) {
	Default.DurationVar(p, full, short, value, usage)
}

func Parse() {
	Default.Parse(os.Args[1:])
}

func Usage() {
	Default.Usage()
}

type Flag struct {
	FullName  string
	ShortName string
	Default   interface{}
	Usage     string
}

func (f *Flag) usageTitle() string {
	var name string
	if f.FullName != "" && f.ShortName != "" {
		name = fmt.Sprintf("-%s, -%s", f.ShortName, f.FullName)
	} else if f.FullName != "" {
		name = fmt.Sprintf("-%s", f.FullName)
	} else {
		name = fmt.Sprintf("-%s", f.ShortName)
	}
	if v := fmt.Sprint(f.Default); v != "" {
		name = name + fmt.Sprintf("[=%v]", v)
	}
	return name
}

type Set struct {
	Desc  string
	flags []*Flag
	inner *flag.FlagSet
	style UsageStyle
}

func NewSet(name, desc string, errorHandling ErrorHandling) *Set {
	return Wrap(flag.NewFlagSet(name, errorHandling), desc)
}

func Wrap(fs *flag.FlagSet, desc string) *Set {
	f := &Set{
		Desc:  desc,
		inner: fs,
		style: Compact,
	}
	f.inner.Usage = f.Usage
	return f
}

func (s *Set) String(full, short, value, usage string) *string {
	p := new(string)
	s.StringVar(p, full, short, value, usage)
	return p
}

func (s *Set) StringVar(p *string, full, short, value, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.StringVar(p, full, value, usage)
	}
	if short != "" {
		s.inner.StringVar(p, short, value, usage)
	}
}

func (s *Set) Int(full, short string, value int, usage string) *int {
	p := new(int)
	s.IntVar(p, full, short, value, usage)
	return p
}

func (s *Set) IntVar(p *int, full, short string, value int, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.IntVar(p, full, value, usage)
	}
	if short != "" {
		s.inner.IntVar(p, short, value, usage)
	}
}

func (s *Set) Int64(full, short string, value int64, usage string) *int64 {
	p := new(int64)
	s.Int64Var(p, full, short, value, usage)
	return p
}

func (s *Set) Int64Var(p *int64, full, short string, value int64, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.Int64Var(p, full, value, usage)
	}
	if short != "" {
		s.inner.Int64Var(p, short, value, usage)
	}
}

func (s *Set) Uint(full, short string, value uint, usage string) *uint {
	p := new(uint)
	s.UintVar(p, full, short, value, usage)
	return p
}

func (s *Set) UintVar(p *uint, full, short string, value uint, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.UintVar(p, full, value, usage)
	}
	if short != "" {
		s.inner.UintVar(p, short, value, usage)
	}
}

func (s *Set) Uint64(full, short string, value uint64, usage string) *uint64 {
	p := new(uint64)
	s.Uint64Var(p, full, short, value, usage)
	return p
}

func (s *Set) Uint64Var(p *uint64, full, short string, value uint64, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.Uint64Var(p, full, value, usage)
	}
	if short != "" {
		s.inner.Uint64Var(p, short, value, usage)
	}
}

func (s *Set) Float64(full, short string, value float64, usage string) *float64 {
	p := new(float64)
	s.Float64Var(p, full, short, value, usage)
	return p
}

func (s *Set) Float64Var(p *float64, full, short string, value float64, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.Float64Var(p, full, value, usage)
	}
	if short != "" {
		s.inner.Float64Var(p, short, value, usage)
	}
}

func (s *Set) Bool(full, short string, value bool, usage string) *bool {
	p := new(bool)
	s.BoolVar(p, full, short, value, usage)
	return p
}

func (s *Set) BoolVar(p *bool, full, short string, value bool, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.BoolVar(p, full, value, usage)
	}
	if short != "" {
		s.inner.BoolVar(p, short, value, usage)
	}
}

func (s *Set) Duration(full, short string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	s.DurationVar(p, full, short, value, usage)
	return p
}

func (s *Set) DurationVar(p *time.Duration, full, short string, value time.Duration, usage string) {
	s.addFlag(full, short, value, usage)
	if full != "" {
		s.inner.DurationVar(p, full, value, usage)
	}
	if short != "" {
		s.inner.DurationVar(p, short, value, usage)
	}
}

// Register adds common flags like help/version/profile/config etc.
func (s *Set) Register(f CommonFlag) {
	if f&Help > 0 {
		s.Bool("help", "h", false, "Show help")
	}
	if f&Version > 0 {
		s.Bool("version", "v", false, "Print version information")
	}
	if f&Profile > 0 {
		s.String("profile", "p", "", "Set active profiles")
	}
	if f&Config > 0 {
		s.String("config", "c", "", "Set configuration directory")
	}
}

func (s *Set) SetUsageStyle(style UsageStyle) {
	s.style = style
}

func (s *Set) addFlag(full, short string, value interface{}, usage string) {
	f := &Flag{
		FullName:  full,
		ShortName: short,
		Default:   value,
		Usage:     usage,
	}
	s.flags = append(s.flags, f)
}

func (s *Set) Parse(args []string) {
	s.inner.Parse(args)
}

// Args returns the non-flag arguments.
func (s *Set) Args() []string {
	return s.inner.Args()
}

func (s *Set) Set(name, value string) error {
	return s.inner.Set(name, value)
}

func (s *Set) Lookup(name string) *flag.Flag {
	return s.inner.Lookup(name)
}

// Inner returns internal `flag.FlagSet` used by Set.
func (s *Set) Inner() *flag.FlagSet {
	return s.inner
}

func (s *Set) Usage() {
	if s.Desc != "" {
		fmt.Print(s.Desc + "\n\n")
	}

	fmt.Print("Options:\n\n")
	if s.style == Compact {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for _, f := range s.flags {
			s.style.Print(w, f.usageTitle(), f.Usage)
		}
		w.Flush()
	} else {
		for i, f := range s.flags {
			s.style.Print(os.Stdout, f.usageTitle(), f.Usage)
			if i != len(s.flags)-1 {
				fmt.Println()
			}
		}
	}
}

package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuigh/auxo"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/log"
)

const PkgName = "auxo.app"

var (
	// Version is the version of application
	Version string
	// SCMRevision is commit hash of source tree
	SCMRevision string
	// SCMBranch is current branch name the code is built off
	SCMBranch string
	// BuildTime is RFC3339 formatted UTC date, e.g. 2017-12-01T13:04:23Z
	BuildTime string
)

var (
	// Timeout is the amount of time allowed to wait graceful shutdown.
	Timeout = time.Second * 30

	initializers []func() error
	closers      []func()
)

type (
	// Server defines the interface of server application.
	Server interface {
		Serve() error
		Close(timeout time.Duration)
	}

	// ServeFunc is server execute method.
	ServeFunc func() error

	// CloseFunc is server shutdown method.
	CloseFunc func(timeout time.Duration)
)

// Path returns program's full filename
func Path() string {
	p, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return p
}

// Ensure exits program and log err if err is not nil.
func Ensure(err error) {
	if err != nil {
		log.Get(PkgName).Fatal("app > ", err)
	}
}

// Run executes program and subscribe exit signals.
func Run(s Server, signals ...os.Signal) {
	RunFunc(s.Serve, s.Close, signals...)
}

// RunFunc executes program and subscribe exit signals.
func RunFunc(runner ServeFunc, closer CloseFunc, signals ...os.Signal) {
	// subscribe signals
	stop := make(chan os.Signal, 1)
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	signal.Notify(stop, signals...)

	logger := log.Get(PkgName)
	go func() {
		if err := runner(); err != nil {
			logger.Fatal("app > ", err)
		}
	}()

	sig := <-stop // wait for signals
	logger.Info("app > Received signal: ", sig.String())
	logger.Info("app > Exiting program...")
	if closer != nil {
		closer(Timeout)
	}
	logger.Info("app > Program exited")
}

// OnInit register an initializer which execute on app start. If fn return an error, the app will panic.
func OnInit(fn func() error) {
	if fn != nil {
		initializers = append(initializers, fn)
	}
}

// OnClose subscribes close events
func OnClose(fn func()) {
	if fn != nil {
		closers = append(closers, fn)
	}
}

func Start() {
	var (
		args = os.Args[1:]
		cmd  = &Command{
			Name:     Name,
			Desc:     Desc,
			Action:   Action,
			Flags:    Flags,
			children: children,
		}
		i int
	)

	cmd.Flags.Desc = Desc

	for l := len(args); i < l; i++ {
		if c := cmd.children[args[i]]; c == nil {
			break
		} else {
			cmd = c
		}
	}

	if cmd.Flags != nil {
		cmd.Flags.Parse(args[i:])
		config.BindFlags(cmd.Flags.Inner())
	}

	// trigger initializers
	for _, fn := range initializers {
		if err := fn(); err != nil {
			panic(err)
		}
	}
	defer func() {
		// trigger closers
		for _, fn := range closers {
			fn()
		}
	}()

	ctx := &Context{cmd: cmd}
	handleCommonFlags(ctx)

	// print banner
	if config.GetBool("banner") {
		fmt.Print(auxo.Banner)
		fmt.Println("\tVERSION " + auxo.Version)
		fmt.Println()
	}

	if cmd.Action != nil {
		if err := cmd.Action(ctx); err != nil {
			log.Get(PkgName).Error(err)
			if e, ok := err.(Error); ok {
				os.Exit(e.Code())
			} else {
				os.Exit(1)
			}
		}
	}
}

func handleCommonFlags(ctx *Context) {
	if ctx.Help() {
		ctx.Usage()
		os.Exit(0)
	}

	if ctx.Version() {
		printVersion()
		os.Exit(0)
	}

	if profiles := ctx.Profiles(); profiles != nil {
		config.SetProfile(profiles...)
	}

	if c := ctx.Config(); c != "" {
		config.AddFolder(c)
	}
}

func printVersion() {
	rev := SCMRevision
	if rev == "" {
		rev = "?"
	}
	fmt.Printf("%s %s (auxo: %s, rev: %s)", Name, Version, auxo.Version, rev)
}

type Error interface {
	error
	Code() int
}

type fatalError struct {
	msg  string
	code int
}

func (e *fatalError) Error() string {
	return e.msg
}

func (e *fatalError) Code() int {
	return e.code
}

func Fatal(code int, msg interface{}) Error {
	return &fatalError{
		code: code,
		msg:  fmt.Sprint(msg),
	}
}

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
	Timeout     = time.Second * 30
	closeEvents []func()
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

// Run executes program and subscribe exit signals.
func Run(s Server, signals ...os.Signal) {
	RunFunc(s.Serve, s.Close, signals...)
}

// RunFunc executes program and subscribe exit signals.
func RunFunc(runner ServeFunc, closer CloseFunc, signals ...os.Signal) {
	// print banner
	if config.GetBool("banner") {
		fmt.Print(auxo.Banner)
		fmt.Println("\tVERSION " + auxo.Version)
		fmt.Println()
	}

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

	for _, fn := range closeEvents {
		fn()
	}

	logger.Info("app > Exiting program...")
	if closer != nil {
		closer(Timeout)
	}
	logger.Info("app > Server stopped")
}

// OnClose subscribes close events
func OnClose(fn func()) {
	if fn != nil {
		closeEvents = append(closeEvents, fn)
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
		i   int
		arg string
	)

	cmd.Flags.Desc = Desc

	for i, arg = range args {
		if c := cmd.children[arg]; c == nil {
			break
		} else {
			cmd = c
		}
	}

	if cmd.Flags != nil {
		cmd.Flags.Parse(args[i:])
		config.BindFlags(cmd.Flags.Inner())
	}
	if cmd.Action != nil {
		ctx := &Context{cmd: cmd}
		handleCommonFlags(ctx)
		cmd.Action(ctx)
	}
}

func handleCommonFlags(ctx *Context) {
	if ctx.Help() {
		ctx.cmd.Flags.Usage()
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

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
	Version  string
	Revision string

	closeEvents []func()
)

type (
	// Program is interface of a program.
	Program interface {
		Run() error
		Close(timeout time.Duration)
	}

	// Runner is program execute method.
	Runner func() error

	// Closer is program shutdown method.
	Closer func(timeout time.Duration)
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
func Run(p Program, signals ...os.Signal) {
	RunFunc(p.Run, p.Close, signals...)
}

// RunFunc executes program and subscribe exit signals.
func RunFunc(runner Runner, closer Closer, signals ...os.Signal) {
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
		closer(time.Second * 30)
	}
	logger.Info("app > Program stopped")
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
	rev := Revision
	if rev == "" {
		rev = "?"
	}
	fmt.Printf("%s %s (auxo: %s, rev: %s)", Name, Version, auxo.Version, rev)
}

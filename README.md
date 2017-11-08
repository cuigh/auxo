# AUXO

**auxo** is an all-in-one Go package for simplifying program development.

## Quick start

Here is a little demo app developed with **auxo**.

```go
package main

import (
	"fmt"

	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/app/flag"
	"github.com/cuigh/auxo/config"
)

func main() {
	app.Name = "test"
	app.Desc = "a test app for auxo"
	app.Action = func(ctx *app.Context) {
		app.RunFunc(runner, nil)
	}
	app.Register(flag.All)

	cmd := app.NewCommand("test", "a test sub command", func(ctx *app.Context) {
		p := config.GetString("p")
		fmt.Println("profile:", p)
	})
	cmd.Register(flag.Profile)
	app.AddCommand(cmd)

	app.Start()
}

func runner() error {
	fmt.Println("Hello, world")
	return nil
}
```

## Log

TODO

## Configuration

TODO

## Web

TODO
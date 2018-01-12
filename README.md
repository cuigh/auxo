# AUXO

**auxo** is an all-in-one Go package for simplifying program development.

> **WARNING**: This package is a work in progress. It's API may still break in backwards-incompatible ways without warnings. Please use dependency management tools such as **[dep](https://github.com/golang/dep)** or **[glide](https://github.com/Masterminds/glide)** to lock version.

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
	app.Flags.Register(flag.All)

	cmd := app.NewCommand("test", "a test sub command", func(ctx *app.Context) {
		p := config.GetString("p")
		fmt.Println("profile:", p)
	})
	cmd.Flags.Register(flag.Profile)
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
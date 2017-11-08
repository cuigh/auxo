package console

import (
	"io"
	"os"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/util/cast"
)

type writer struct {
	colorized bool
	io.Writer
}

func New(options data.Map) (io.Writer, error) {
	return &writer{
		colorized: cast.ToBool(options.Get("colorized")),
		Writer:    os.Stdout,
	}, nil
}

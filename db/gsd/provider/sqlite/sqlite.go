package sqlite

import (
	"strconv"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/db/gsd/provider"
	"github.com/cuigh/auxo/errors"
)

func quote(b *gsd.Builder, s string) {
	b.AppendByte('[')
	b.Append(s)
	b.AppendByte(']')
}

func limit(b *gsd.Builder, skip, take int) {
	b.Append(" LIMIT ", strconv.Itoa(take), " OFFSET ", strconv.Itoa(skip))
}

func call(b *gsd.Builder, sp string, args ...interface{}) error {
	return errors.NotSupported
}

func New(_ data.Map) gsd.Provider {
	return &provider.Provider{
		Quote: quote,
		Limit: limit,
		Call:  call,
	}
}

func init() {
	gsd.RegisterProvider("sqlite", New)
}

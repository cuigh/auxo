package mysql

import (
	"strconv"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/db/gsd/provider"
)

func quote(b *gsd.Builder, s string) {
	b.AppendByte('`')
	b.Append(s)
	b.AppendByte('`')
}

func limit(b *gsd.Builder, skip, take int) {
	b.Append(" LIMIT ", strconv.Itoa(skip), ",", strconv.Itoa(take))
}

func call(b *gsd.Builder, sp string, args ...interface{}) (err error) {
	//CALL SP(?,?,?)
	b.Append("CALL ", sp, "(")
	for i := range args {
		if i > 0 {
			b.AppendByte(',')
		}
		b.AppendByte('?')
	}
	b.AppendByte(')')
	b.Args = args
	return
}

func New(_ data.Map) gsd.Provider {
	return &provider.Provider{
		Quote: quote,
		Limit: limit,
		Call:  call,
	}
}

func init() {
	gsd.RegisterProvider("mysql", New)
}

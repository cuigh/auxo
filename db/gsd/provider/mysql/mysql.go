package mysql

import (
	"strconv"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/db/gsd/provider"
)

func quote(b *gsd.Builder, s string) {
	b.WriteByte('`').WriteString(s).WriteByte('`')
}

func limit(b *gsd.Builder, skip, take int) {
	b.WriteString(" LIMIT ", strconv.Itoa(skip), ",", strconv.Itoa(take))
}

func call(b *gsd.Builder, sp string, args ...interface{}) (err error) {
	//CALL SP(?,?,?)
	b.WriteString("CALL ", sp, "(")
	for i := range args {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	b.WriteByte(')')
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

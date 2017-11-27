package mssql

import (
	"strconv"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/db/gsd/provider"
)

func quote(b *gsd.Builder, s string) {
	b.AppendByte('[')
	b.Append(s)
	b.AppendByte(']')
}

func limit(b *gsd.Builder, skip, take int) {
	b.Append(" OFFSET ", strconv.Itoa(skip), " ROWS FETCH NEXT ", strconv.Itoa(take), " ROWS ONLY")
}

func call(b *gsd.Builder, sp string, args ...interface{}) (err error) {
	//SET NOCOUNT ON; EXEC [SP] ?; SET NOCOUNT OFF
	b.Append("SET NOCOUNT ON; EXEC [", sp, "]")
	for i := range args {
		if i > 0 {
			b.AppendByte(',')
		}
		b.AppendByte('?')
	}
	b.Append("SET NOCOUNT OFF")
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
	gsd.RegisterProvider("mssql", New)
}

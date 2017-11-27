package abbr

import (
	"strings"

	"github.com/cuigh/auxo/db/gsd"
)

func ASC(cols ...string) *gsd.Order {
	return &gsd.Order{
		Columns: *C(cols...),
		Type:    gsd.ASC,
	}
}

func DESC(cols ...string) *gsd.Order {
	return &gsd.Order{
		Columns: *C(cols...),
		Type:    gsd.DESC,
	}
}

func T(name string, alias ...string) gsd.Table {
	return gsd.NewTable(name, alias...)
}

func C(cols ...string) *gsd.Columns {
	return gsd.NewColumns(cols...)
}

func CT(table interface{}, cols ...string) *gsd.Columns {
	return gsd.NewColumns().Table(table, cols...)
}

func CX(expr, alias string) *gsd.Columns {
	return gsd.NewColumns().Expr(expr, alias)
}

func CS(cols string) *gsd.Columns {
	return gsd.NewColumns(strings.Split(cols, ",")...)
}

func Omit(cols ...string) gsd.ColumnFilter {
	return gsd.Omit(cols...)
}

func Include(cols ...string) gsd.ColumnFilter {
	return gsd.Include(cols...)
}

func F() *gsd.SimpleFilters {
	return gsd.NewFilters()
}

func Not(inner gsd.Filters) gsd.Filters {
	return gsd.Not(inner)
}

func And(left, right gsd.Filters) gsd.Filters {
	return gsd.And(left, right)
}

func Or(left, right gsd.Filters) gsd.Filters {
	return gsd.Or(left, right)
}

func Equal(col, val interface{}) *gsd.SimpleFilters {
	return gsd.Equal(col, val)
}

func On(left, right interface{}) *gsd.SimpleFilters {
	return gsd.NewFilters().Equal2(left, right)
}

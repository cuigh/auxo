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

func W() *gsd.SimpleCriteriaSet {
	return &gsd.SimpleCriteriaSet{}
}

func Not(inner gsd.CriteriaSet) gsd.CriteriaSet {
	return gsd.Not(inner)
}

func And(left, right gsd.CriteriaSet) gsd.CriteriaSet {
	return gsd.And(left, right)
}

func Or(left, right gsd.CriteriaSet) gsd.CriteriaSet {
	return gsd.Or(left, right)
}

func Equal(col, val interface{}) *gsd.SimpleCriteriaSet {
	return gsd.Equal(col, val)
}

func On(left, right interface{}) *gsd.SimpleCriteriaSet {
	return W().Equal2(left, right)
}

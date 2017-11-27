package gsd

import "github.com/cuigh/auxo/ext/texts"

var countColumns = Columns{NewExprColumn("COUNT(0)", "count")}

type ColumnFilter func(cols []string) []string

func Include(cols ...string) ColumnFilter {
	return func(raw []string) []string {
		return cols
	}
}

func Omit(cols ...string) ColumnFilter {
	return func(raw []string) (c []string) {
		for _, v := range raw {
			if !texts.Contains(cols, v) {
				c = append(c, v)
			}
		}
		return
	}
}

type Columns []Column

func NewColumns(cols ...string) *Columns {
	c := &Columns{}
	for _, col := range cols {
		*c = append(*c, SimpleColumn(col))
	}
	return c
}

func (c *Columns) Simple(cols ...string) *Columns {
	for _, col := range cols {
		*c = append(*c, SimpleColumn(col))
	}
	return c
}

func (c *Columns) Table(t interface{}, cols ...string) *Columns {
	for _, col := range cols {
		*c = append(*c, &TableColumn{t: toTable(t), name: col})
	}
	return c
}

func (c *Columns) Alias(t Table, col, alias string) *Columns {
	*c = append(*c, &TableColumn{t: t, name: col, alias: alias})
	return c
}

func (c *Columns) Expr(expr, alias string) *Columns {
	*c = append(*c, &ExprColumn{expr: expr, alias: alias})
	return c
}

func (c *Columns) ASC() *Order {
	return &Order{Columns: *c, Type: ASC}
}

func (c *Columns) DESC() *Order {
	return &Order{Columns: *c, Type: DESC}
}

type Column interface {
	Table() Table
	Name() string
	Alias() string
	Field() string
}

type SimpleColumn string

func NewColumn(name string, alias ...string) Column {
	if len(alias) > 0 {
		return NewTableColumn(nil, name, alias...)
	}
	return SimpleColumn(name)
}

func (SimpleColumn) Table() Table {
	return nil
}

func (c SimpleColumn) Name() string {
	return string(c)
}

func (c SimpleColumn) Alias() string {
	return ""
}

func (c SimpleColumn) Field() string {
	return string(c)
}

type TableColumn struct {
	t     Table
	name  string
	alias string
}

func NewTableColumn(table interface{}, name string, alias ...string) Column {
	c := &TableColumn{
		t:    toTable(table),
		name: name,
	}
	if len(alias) > 0 {
		c.alias = alias[0]
	}
	return c
}

func (c *TableColumn) Table() Table {
	return c.t
}

func (c *TableColumn) Name() string {
	return c.name
}

func (c *TableColumn) Alias() string {
	return c.alias
}

func (c *TableColumn) Field() string {
	if c.alias == "" {
		return c.name
	}
	return c.alias
}

type ExprColumn struct {
	expr  string
	alias string
}

func NewExprColumn(expr, alias string) Column {
	return &ExprColumn{expr: expr, alias: alias}
}

func (c *ExprColumn) Table() Table {
	return nil
}

func (c *ExprColumn) Name() string {
	return c.expr
}

func (c *ExprColumn) Alias() string {
	return c.alias
}

func (c *ExprColumn) Field() string {
	return c.alias
}

//type funcColumn struct {
//	fn    string
//	t     Table
//	name  string
//	alias string
//}
//
//func (c *funcColumn) Table() Table {
//	return c.t
//}
//
//func (c *funcColumn) Name() string {
//	return c.fn
//}
//
//func (c *funcColumn) Alias() string {
//	return c.alias
//}
//
//func (c *funcColumn) Field() string {
//	return c.alias
//}

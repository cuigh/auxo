package gsd

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
)

//type LockMode int
//
//const (
//	LockNone LockMode = iota
//	LockShared
//	LockExclusive
//)

type OrderType int

const (
	ASC OrderType = iota
	DESC
)

type Order struct {
	Type OrderType
	Columns
}

type SelectInfo struct {
	Table
	Distinct bool
	Columns
	Where CriteriaSet
	Joins []struct {
		Type string
		Table
		On CriteriaSet
	}
	Groups Columns
	Having CriteriaSet
	Orders []*Order
	Skip   int
	Take   int
	Count  bool // return total count
}

type selectContext struct {
	Builder
	info SelectInfo
	db   *database
	Executor
	context.Context
}

func (c *selectContext) Reset() {
	c.info.Distinct = false
	c.info.Columns = nil
	c.info.Where = nil
	c.info.Joins = nil
	c.info.Groups = nil
	c.info.Having = nil
	c.info.Orders = nil
	c.info.Skip, c.info.Take = 0, 0
	c.info.Count = false
	c.Builder.Reset()
}

func (c *selectContext) Select(cols *Columns, distinct ...bool) SelectClause {
	c.info.Columns = *cols
	if len(distinct) > 0 {
		c.info.Distinct = distinct[0]
	}
	return c
}

func (c *selectContext) Count(table interface{}) FromClause {
	c.info.Columns = Columns{NewExprColumn("COUNT(0)", "count")}
	c.info.Table = toTable(table)
	return c
}

func (c *selectContext) From(table interface{}) FromClause {
	c.info.Table = toTable(table)
	return c
}

func (c *selectContext) Where(w CriteriaSet) WhereClause {
	c.info.Where = w
	return c
}

func (c *selectContext) Join(t interface{}, on CriteriaSet) JoinClause {
	return c.join(t, on, "JOIN")
}

func (c *selectContext) LeftJoin(t interface{}, on CriteriaSet) JoinClause {
	return c.join(t, on, "LEFT JOIN")
}

func (c *selectContext) RightJoin(t interface{}, on CriteriaSet) JoinClause {
	return c.join(t, on, "RIGHT JOIN")
}

func (c *selectContext) FullJoin(t interface{}, on CriteriaSet) JoinClause {
	return c.join(t, on, "FULL JOIN")
}

func (c *selectContext) join(t interface{}, on CriteriaSet, jt string) JoinClause {
	c.info.Joins = append(c.info.Joins, struct {
		Type string
		Table
		On CriteriaSet
	}{Type: jt, Table: toTable(t), On: on})
	return c
}

func (c *selectContext) GroupBy(cols *Columns) GroupByClause {
	c.info.Groups = *cols
	return c
}

func (c *selectContext) Having(f CriteriaSet) HavingClause {
	c.info.Having = f
	return c
}

func (c *selectContext) OrderBy(orders ...*Order) OrderByClause {
	c.info.Orders = orders
	return c
}

func (c *selectContext) Limit(skip, take int) SelectResultClause {
	c.info.Skip, c.info.Take = skip, take
	return c
}

func (c *selectContext) Page(index, size int) SelectResultClause {
	c.info.Skip, c.info.Take = (index-1)*size, size
	return c
}

//func (c *selectContext) One(i interface{}) error {
//	m := GetMeta(reflect.TypeOf(i))
//	cols := make([]string, 0, len(c.info.Columns))
//	for i := 0; i < len(c.info.Columns); i++ {
//		cols = append(cols, c.info.Columns[i].Field())
//	}
//
//	row, err := c.row()
//	if err != nil {
//		return err
//	}
//	values := m.Pointers(i, cols...)
//	return row.Scan(values...)
//}

func (c *selectContext) Scan(values ...interface{}) error {
	row, err := c.row()
	if err != nil {
		return err
	}
	return row.Scan(values...)
}

func (c *selectContext) Int() (i int, err error) {
	var row *sql.Row
	if row, err = c.row(); err == nil {
		err = row.Scan(&i)
	}
	return
}

func (c *selectContext) Value() (v Value) {
	var row *sql.Row
	if row, v.err = c.row(); v.err == nil {
		v.err = row.Scan(&v.bytes)
	}
	return
}

func (c *selectContext) Fill(i interface{}) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	return r.Fill(i)
}

func (c *selectContext) Reader() (Reader, error) {
	rows, err := c.rows()
	return (*reader)(rows), err
}

func (c *selectContext) For(fn func(r Reader) error) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}

func (c *selectContext) List(i interface{}, total *int) error {
	// i = &[]struct{}/&[]*struct{}
	c.info.Count = true

	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	if err = r.Fill(i); err != nil {
		return err
	}

	if r.NextSet() && r.Next() {
		return r.Scan(total)
	} else {
		return errors.New("gsd: can't read count")
	}
}

func (c *selectContext) Load(i interface{}) error {
	t := reflect.TypeOf(i)
	m := GetMeta(t)

	// Columns
	c.info.Table = NewTable(m.Table)
	c.info.Columns = make(Columns, len(m.Selects))
	for i, col := range m.Selects {
		c.info.Columns[i] = SimpleColumn(col)
	}

	// Where
	keys := m.PrimaryKeys
	values := m.PrimaryKeyValues(i)
	where := &SimpleCriteriaSet{}
	for i, v := range values {
		where.Equal(keys[i], v)
	}
	c.info.Where = where

	// Exec
	row, err := c.row()
	if err != nil {
		return err
	}
	values = m.Pointers(i, m.Selects...)
	return row.Scan(values...)
}

func (c *selectContext) row() (row *sql.Row, err error) {
	defer ctxPool.PutSelect(c)

	if err = c.db.p.BuildSelect(&c.Builder, &c.info); err != nil {
		return
	}
	return c.QueryRow(c.Context, c.Builder.String(), c.Builder.Args...), nil
}

func (c *selectContext) rows() (rows *sql.Rows, err error) {
	defer ctxPool.PutSelect(c)

	if err = c.db.p.BuildSelect(&c.Builder, &c.info); err != nil {
		return
	}
	return c.QueryRows(c.Context, c.Builder.String(), c.Builder.Args...)
}

type countContext selectContext

func (c *countContext) Count(table interface{}) CountClause {
	c.info.Columns = countColumns
	c.info.Table = toTable(table)
	return c
}

func (c *countContext) Join(t interface{}, on CriteriaSet) CountClause {
	(*selectContext)(c).Join(t, on)
	return c
}

func (c *countContext) LeftJoin(t interface{}, on CriteriaSet) CountClause {
	(*selectContext)(c).LeftJoin(t, on)
	return c
}

func (c *countContext) RightJoin(t interface{}, on CriteriaSet) CountClause {
	(*selectContext)(c).RightJoin(t, on)
	return c
}

func (c *countContext) FullJoin(t interface{}, on CriteriaSet) CountClause {
	(*selectContext)(c).FullJoin(t, on)
	return c
}

func (c *countContext) Where(f CriteriaSet) CountWhereClause {
	(*selectContext)(c).Where(f)
	return c
}

func (c *countContext) GroupBy(cols *Columns) CountGroupByClause {
	(*selectContext)(c).GroupBy(cols)
	return c
}
func (c *countContext) Having(f CriteriaSet) CountResultClause {
	(*selectContext)(c).Having(f)
	return c
}

func (c *countContext) Value() (int, error) {
	return (*selectContext)(c).Int()
}

func (c *countContext) Scan(dst interface{}) error {
	return (*selectContext)(c).Scan(dst)
}

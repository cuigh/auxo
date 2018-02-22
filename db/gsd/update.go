package gsd

import (
	"context"
	"reflect"
)

type updateValue struct {
	Value interface{}
}
type IncValue *updateValue
type DecValue *updateValue
type ExprValue string

type UpdateInfo struct {
	Table   string
	Columns []string
	Values  []interface{}
	Where   CriteriaSet
	Filter  ColumnFilter
}

type updateContext struct {
	Builder
	info UpdateInfo
	db   *database
	Executor
	context.Context
}

func (c *updateContext) Reset() {
	c.info.Columns = nil
	c.info.Values = nil
	c.info.Where = nil
	c.info.Filter = nil
	c.Builder.Reset()
}

func (c *updateContext) Update(table string) UpdateClause {
	c.info.Table = table
	return c
}

func (c *updateContext) Set(col string, val interface{}) SetClause {
	c.info.Columns = append(c.info.Columns, col)
	c.info.Values = append(c.info.Values, val)
	return c
}

func (c *updateContext) Inc(col string, val interface{}) SetClause {
	c.info.Columns = append(c.info.Columns, col)
	c.info.Values = append(c.info.Values, IncValue(&updateValue{val}))
	return c
}

func (c *updateContext) Dec(col string, val interface{}) SetClause {
	c.info.Columns = append(c.info.Columns, col)
	c.info.Values = append(c.info.Values, DecValue(&updateValue{val}))
	return c
}

func (c *updateContext) Expr(col string, val string) SetClause {
	c.info.Columns = append(c.info.Columns, col)
	c.info.Values = append(c.info.Values, ExprValue(val))
	return c
}

func (c *updateContext) Where(w CriteriaSet) ResultClause {
	c.info.Where = w
	return c
}

func (c *updateContext) Modify(i interface{}, filter ...ColumnFilter) (r Result, err error) {
	if len(filter) > 0 {
		c.info.Filter = filter[0]
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("gsd: Modify of non-struct Type " + t.String())
	}

	m := GetMeta(t)
	c.info.Table = m.Table

	// columns and values
	if c.info.Filter == nil {
		c.info.Columns = m.Updates
		c.info.Values = m.UpdateValues(i)
	} else {
		c.info.Columns = c.info.Filter(m.Updates)
		c.info.Values = m.Values(i, c.info.Columns...)
	}

	// filters
	where := &SimpleCriteriaSet{}
	values := m.PrimaryKeyValues(i)
	for i, key := range m.PrimaryKeys {
		where.Equal(key, values[i])
	}
	c.info.Where = where
	return c.Result()
}

func (c *updateContext) Result() (r Result, err error) {
	defer ctxPool.PutUpdate(c)

	if err = c.db.p.BuildUpdate(&c.Builder, &c.info); err != nil {
		return
	}
	return c.Exec(c.Context, c.Builder.String(), c.Builder.Args...)
}

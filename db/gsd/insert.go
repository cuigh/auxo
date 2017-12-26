package gsd

import (
	"context"
	"reflect"
)

type InsertInfo struct {
	Table   string
	Columns []string
	Values  []interface{} // single row or multiple rows
	Filter  ColumnFilter
}

type insertContext struct {
	Builder
	info InsertInfo
	db   *database
	Executor
	context.Context
}

func (c *insertContext) Reset() {
	c.info.Columns = nil
	c.info.Values = nil
	c.info.Filter = nil
	c.Builder.Reset()
}

func (c *insertContext) Insert(table string) InsertClause {
	c.info.Table = table
	return c
}

func (c *insertContext) Create(i interface{}, filter ...ColumnFilter) error {
	if len(filter) > 0 {
		c.info.Filter = filter[0]
	}

	var m *Meta
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t := t.Elem()
		m = GetMeta(t)
		c.info.Table = m.Table

		if c.info.Filter == nil {
			c.info.Columns = m.Inserts
			v := reflect.ValueOf(i)
			rows := v.Len()
			c.info.Values = make([]interface{}, len(c.info.Columns)*rows)
			for i := 0; i < rows; i++ {
				m.CopyInsertValues(v.Index(i).Interface(), c.info.Values[len(c.info.Columns)*i:])
			}
		} else {
			c.info.Columns = c.info.Filter(m.Inserts)
			v := reflect.ValueOf(i)
			rows := v.Len()
			c.info.Values = make([]interface{}, len(m.Columns)*rows)
			for i := 0; i < rows; i++ {
				m.CopyValues(v.Index(i).Interface(), c.info.Values[len(c.info.Columns)*i:], c.info.Columns...)
			}
		}

		return c.Submit()
	} else {
		m = GetMeta(t)
		c.info.Table = m.Table
		if c.info.Filter == nil {
			c.info.Columns = m.Inserts
			c.info.Values = m.InsertValues(i)
		} else {
			c.info.Columns = c.info.Filter(m.Inserts)
			c.info.Values = m.Values(i, c.info.Columns...)
		}

		// set auto increment id
		r, err := c.Result()
		if err != nil {
			return err
		}
		if m.Auto != "" {
			if id, err := r.LastInsertId(); err == nil {
				m.SetAutoValue(i, id)
			} else {
				return err
			}
		}
		return nil
	}
}

func (c *insertContext) Columns(cols ...string) InsertColumnsClause {
	c.info.Columns = cols
	return c
}

func (c *insertContext) Values(values ...interface{}) InsertValuesClause {
	c.info.Values = append(c.info.Values, values...)
	return c
}

func (c *insertContext) Submit() (err error) {
	_, err = c.Result()
	return
}

func (c *insertContext) Result() (r InsertResult, err error) {
	defer ctxPool.PutInsert(c)

	if err = c.db.p.BuildInsert(&c.Builder, &c.info); err != nil {
		return
	}
	return c.Exec(c.Context, c.Builder.String(), c.Builder.Args...)
}

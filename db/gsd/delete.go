package gsd

import (
	"context"
	"reflect"
)

type DeleteInfo struct {
	Table string
	Where Filters
}

type deleteContext struct {
	Builder
	info DeleteInfo
	db   *database
	Executor
	context.Context
}

func (c *deleteContext) Reset() {
	c.info.Where = nil
	c.Builder.Reset()
}

func (c *deleteContext) Delete(table string) DeleteClause {
	c.info.Table = table
	return c
}

func (c *deleteContext) Where(w Filters) ResultClause {
	c.info.Where = w
	return c
}

func (c *deleteContext) Remove(i interface{}) ResultClause {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("gsd: Remove of non-struct Type " + t.String())
	}

	m := GetMeta(t)
	c.info.Table = m.Table

	where := NewFilters()
	values := m.PrimaryKeyValues(i)
	for i, key := range m.PrimaryKeys {
		where.Equal(key, values[i])
	}
	c.info.Where = where
	return c
}

func (c *deleteContext) Submit() (err error) {
	_, err = c.Result()
	return
}

func (c *deleteContext) Result() (r Result, err error) {
	defer ctxPool.PutDelete(c)

	err = c.db.p.BuildDelete(&c.Builder, &c.info)
	if err != nil {
		return
	}
	return c.Exec(c, c.Builder.String(), c.Builder.Args...)
}

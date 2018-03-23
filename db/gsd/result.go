package gsd

import (
	"database/sql"
	"reflect"

	"github.com/cuigh/auxo/errors"
)

type Result interface {
	RowsAffected() (int64, error)
}

type InsertResult interface {
	Result
	LastInsertId() (int64, error)
}

type ExecuteResult interface {
	Result
	LastInsertId() (int64, error)
	//Value() *Value
}

type Reader interface {
	Scan(dst ...interface{}) error
	Fill(i interface{}) error
	Next() bool
	NextSet() bool
	Close() error
	Err() error
}

type reader sql.Rows

func (r *reader) Scan(dst ...interface{}) error {
	return (*sql.Rows)(r).Scan(dst...)
}

func (r *reader) Fill(i interface{}) error {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		switch t.Kind() {
		case reflect.Struct:
			return r.fillOne(i, t)
		case reflect.Slice:
			return r.fillList(i, t)
		}
	}
	return errors.New("gsd: Filling target must be struct-pointer or slice-pointer of struct")
}

func (r *reader) fillOne(i interface{}, t reflect.Type) error {
	rows := (*sql.Rows)(r)
	if !rows.Next() {
		return ErrNoRows
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	m := GetMeta(t)
	values := m.Pointers(i, cols...)
	return rows.Scan(values...)
}

func (r *reader) fillList(i interface{}, t reflect.Type) error {
	rows := (*sql.Rows)(r)
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var elemIsPtr bool
	t = t.Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		elemIsPtr = true
	}
	m := GetMeta(t)
	sv := reflect.ValueOf(i).Elem()
	slice := sv.Slice(0, 0)
	for rows.Next() {
		v := reflect.New(t)
		values := m.Pointers(v.Interface(), cols...)
		err := rows.Scan(values...)
		if err != nil {
			return err
		}
		if !elemIsPtr {
			v = v.Elem()
		}
		slice = reflect.Append(slice, v)
	}
	if rows.Err() == nil {
		sv.Set(slice)
	}
	return rows.Err()
}

func (r *reader) Next() bool {
	return (*sql.Rows)(r).Next()
}

func (r *reader) NextSet() bool {
	return (*sql.Rows)(r).NextResultSet()
}

func (r *reader) Close() error {
	return (*sql.Rows)(r).Close()
}

func (r *reader) Err() error {
	return (*sql.Rows)(r).Err()
}

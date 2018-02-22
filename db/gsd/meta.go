package gsd

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/ext/texts"
)

var (
	typeNullTime    = reflect.TypeOf(NullTime{})
	typeNullBool    = reflect.TypeOf(sql.NullBool{})
	typeNullString  = reflect.TypeOf(sql.NullString{})
	typeNullInt32   = reflect.TypeOf(NullInt32{})
	typeNullInt64   = reflect.TypeOf(sql.NullInt64{})
	typeNullFloat32 = reflect.TypeOf(NullFloat32{})
	typeNullFloat64 = reflect.TypeOf(sql.NullFloat64{})
)

//type TestEntity struct {
//	ID         int32 `gsd:"id,auto"`
//	Name       int32 `gsd:"name,key,size:20"`
//	Code       int32 `gsd:",key"`
//	CreateTime int32 `gsd:",key:false,update:false,insert:true,select:true"`
//	Ignore     int32 `gsd:"-"`
//}

var (
	metaFactory         = &MetaFactory{metas: make(map[reflect.Type]*Meta)}
	DefaultTableOptions = &TableOptions{
		NameStyle:       texts.Lower,
		ColumnNameStyle: texts.Lower,
	}
)

type TableOptions struct {
	Name            string
	NameStyle       texts.NameStyle
	ColumnNameStyle texts.NameStyle
}

// ColumnOptions is table column options
type ColumnOptions struct {
	Name          string `json:"name"`
	AutoIncrement bool   `json:"auto"`
	PrimaryKey    bool   `json:"pk"`
	ForeignKey    bool   `json:"fk"`
	Insert        bool   `json:"insert"`
	Update        bool   `json:"update"`
	Select        bool   `json:"select"`
	Where         *struct {
		Column string
		Type   int
		Index  int
	}
	Size int `json:"size"`
}

type MetaFactory struct {
	locker sync.Mutex
	metas  map[reflect.Type]*Meta
	id     int32
}

func RegisterType(i interface{}, options *TableOptions) {
	if options == nil {
		panic("gsd: RegisterType with nil options")
	}
	t := reflect.TypeOf(i)
	m := newMeta(t, options)
	metaFactory.SetMeta(t, m)
}

func GetMeta(t reflect.Type) *Meta {
	return metaFactory.GetMeta(t)
}

func (f *MetaFactory) GetMeta(t reflect.Type) *Meta {
	f.locker.Lock()
	defer f.locker.Unlock()

	m := f.metas[t]
	if m == nil {
		m = newMeta(t, DefaultTableOptions)
		m.ID = atomic.AddInt32(&f.id, 1) << 8
		f.metas[t] = m
	}
	return m
}

func (f *MetaFactory) SetMeta(t reflect.Type, m *Meta) {
	f.locker.Lock()
	defer f.locker.Unlock()

	m.ID = atomic.AddInt32(&f.id, 1) << 8
	f.metas[t] = m
}

type Meta struct {
	ID          int32
	Table       string
	Auto        string   // auto increment column
	PrimaryKeys []string // primary key columns
	//ForeignKeys   []string // foreign key columns
	Updates []string // update columns
	Inserts []string // insert columns
	Selects []string // select columns
	Columns []string // all columns, columns = keys + updates = autos + inserts
	Wheres  []*struct {
		Column string
		Type   int
		Index  int
	}
	fields        []*reflects.FieldInfo
	fieldMap      map[string]*reflects.FieldInfo
	autoIndex     int   // auto increment column index
	keyIndexes    []int // key column indexes of fields
	insertIndexes []int // insert column indexes of fields
	updateIndexes []int // update column indexes of fields
	selectIndexes []int // select column indexes of fields
}

func newMeta(t reflect.Type, options *TableOptions) *Meta {
	m := &Meta{}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if options.Name != "" {
		m.Table = options.Name
	} else {
		m.Table = texts.Rename(t.Name(), options.NameStyle)
	}

	n := t.NumField()
	m.fieldMap = make(map[string]*reflects.FieldInfo)
	for i := 0; i < n; i++ {
		sf := t.Field(i)
		options := m.parseOptions(&sf, options.ColumnNameStyle)
		if options.Name == "-" { // transient column
			continue
		}

		fi := reflects.NewFieldInfo(&sf)
		m.fieldMap[options.Name] = fi
		m.fields = append(m.fields, fi)
		m.Columns = append(m.Columns, options.Name)
		if options.AutoIncrement {
			if m.Auto == "" {
				m.Auto = options.Name
				m.autoIndex = len(m.fields) - 1
			} else {
				panic(errors.Format("gsd: found multiple auto increment columns: [%s, %s]", m.Auto, options.Name))
			}
		}
		if options.PrimaryKey {
			m.PrimaryKeys = append(m.PrimaryKeys, options.Name)
			m.keyIndexes = append(m.keyIndexes, len(m.fields)-1)
		}
		if options.Insert {
			m.Inserts = append(m.Inserts, options.Name)
			m.insertIndexes = append(m.insertIndexes, len(m.fields)-1)
		}
		if options.Update {
			m.Updates = append(m.Updates, options.Name)
			m.updateIndexes = append(m.updateIndexes, len(m.fields)-1)
		}
		if options.Select {
			m.Selects = append(m.Selects, options.Name)
			m.selectIndexes = append(m.selectIndexes, len(m.fields)-1)
		}
		if options.Where != nil {
			options.Where.Index = len(m.fields) - 1
			m.Wheres = append(m.Wheres, options.Where)
		}
	}
	return m
}

func (m *Meta) parseOptions(sf *reflect.StructField, ns texts.NameStyle) (options *ColumnOptions) {
	options = &ColumnOptions{
		Name:   texts.Rename(sf.Name, ns),
		Insert: true,
		Update: true,
		Select: true,
	}

	// `gsd:"code,key:false,update:false,insert:true,select:true"`
	tag := sf.Tag.Get("gsd")
	if tag == "" {
		return
	}

	items := strings.Split(tag, ",")
	if items[0] != "" {
		options.Name = items[0]
	}

	d := make(map[string]string)
	for i := 1; i < len(items); i++ {
		kv := strings.SplitN(items[i], ":", 2)
		if len(kv) == 1 {
			d[kv[0]] = ""
		} else {
			d[kv[0]] = kv[1]
		}
	}
	if v, ok := d["auto"]; ok {
		options.AutoIncrement = v == "" || v == "true"
		if options.AutoIncrement {
			options.PrimaryKey = true
			options.Insert = false
			options.Update = false
		}
	}
	if v, ok := d["pk"]; ok {
		options.PrimaryKey = v == "" || v == "true"
		if options.PrimaryKey {
			options.Update = false
		}
	}
	if v, ok := d["fk"]; ok {
		options.ForeignKey = v == "" || v == "true"
	}
	if v, ok := d["insert"]; ok {
		options.Insert = v == "" || v == "true"
	}
	if v, ok := d["update"]; ok {
		options.Update = v == "" || v == "true"
	}
	if v, ok := d["select"]; ok {
		options.Select = v == "" || v == "true"
	}
	if v, ok := d["where"]; ok {
		options.Where = &struct {
			Column string
			Type   int
			Index  int
		}{Column: options.Name, Type: parseCriteriaType(v)}
	}
	return
}

func (m *Meta) Value(i interface{}, col string) interface{} {
	fi := m.fieldMap[col]
	return fi.Get(reflects.Ptr(i))
}

func (m *Meta) SetValue(i interface{}, col string, val interface{}) {
	fi := m.fieldMap[col]
	fi.Set(reflects.Ptr(i), val)
}

// Values return fields values
func (m *Meta) Values(i interface{}, cols ...string) []interface{} {
	ptr := reflects.Ptr(i)
	values := make([]interface{}, len(cols))
	for i, f := range cols {
		values[i] = m.fieldMap[f].Get(ptr)
	}
	return values
}

// Pointers return pointers of fields for scanning
func (m *Meta) Pointers(i interface{}, cols ...string) []interface{} {
	ptr := reflects.Ptr(i)
	values := make([]interface{}, len(cols))
	for i, f := range cols {
		if fi := m.fieldMap[f]; fi != nil {
			values[i] = fi.GetPointer(ptr)
		} else {
			var val interface{}
			values[i] = &val
		}
	}
	return values
}

//func (m *Meta) AutoValue(i interface{}) interface{} {
//	ptr := reflects.Ptr(i)
//	fi := m.fields[m.autoIndex]
//	return fi.Get(ptr)
//}

// SetAutoValue set value of auto increment field
func (m *Meta) SetAutoValue(i interface{}, value int64) {
	ptr := reflects.Ptr(i)
	fi := m.fields[m.autoIndex]
	var v interface{}
	switch fi.Type {
	case reflects.TypeInt:
		v = int(value)
	case reflects.TypeInt8:
		v = int8(value)
	case reflects.TypeInt16:
		v = int16(value)
	case reflects.TypeInt32:
		v = int32(value)
	case reflects.TypeInt64:
		v = int64(value)
	case reflects.TypeUint:
		v = uint(value)
	case reflects.TypeUint8:
		v = uint8(value)
	case reflects.TypeUint16:
		v = uint16(value)
	case reflects.TypeUint32:
		v = uint32(value)
	case reflects.TypeUint64:
		v = uint64(value)
	}
	fi.Set(ptr, v)
}

// CopyValues copy fields to struct i
func (m *Meta) CopyValues(i interface{}, values []interface{}, cols ...string) {
	ptr := reflects.Ptr(i)
	for i, f := range cols {
		values[i] = m.fieldMap[f].Get(ptr)
	}
}

// InsertValues return values of insert-fields
func (m *Meta) InsertValues(i interface{}) []interface{} {
	values := make([]interface{}, len(m.insertIndexes))
	m.CopyInsertValues(i, values)
	return values
}

// CopyInsertValues copy values of insert-fields to struct i
func (m *Meta) CopyInsertValues(i interface{}, values []interface{}) {
	ptr := reflects.Ptr(i)
	for vi, fi := range m.insertIndexes {
		values[vi] = m.fields[fi].Get(ptr)
	}
}

// UpdateValues return values of update-fields
func (m *Meta) UpdateValues(i interface{}) []interface{} {
	values := make([]interface{}, len(m.updateIndexes))
	m.CopyUpdateValues(i, values)
	return values
}

// CopyUpdateValues copy values of update-fields to struct i
func (m *Meta) CopyUpdateValues(i interface{}, values []interface{}) {
	ptr := reflects.Ptr(i)
	for vi, fi := range m.updateIndexes {
		values[vi] = m.fields[fi].Get(ptr)
	}
}

// PrimaryKeyValues return values of key-fields
func (m *Meta) PrimaryKeyValues(i interface{}) []interface{} {
	values := make([]interface{}, len(m.keyIndexes))
	m.CopyPrimaryKeyValues(i, values)
	return values
}

// CopyPrimaryKeyValues copy values of pk-fields to struct i
func (m *Meta) CopyPrimaryKeyValues(i interface{}, values []interface{}) {
	ptr := reflects.Ptr(i)
	for vi, fi := range m.keyIndexes {
		values[vi] = m.fields[fi].Get(ptr)
	}
}

// WhereValues return values of searching criteria set.
func (m *Meta) WhereValues(i interface{}) CriteriaSet {
	ptr := reflects.Ptr(i)
	cs := &SimpleCriteriaSet{}
	for _, item := range m.Wheres {
		cs.add(item.Column, item.Type, m.fields[item.Index].Get(ptr))
	}
	return cs
}

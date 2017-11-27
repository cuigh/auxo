package gsd_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/test/assert"
)

func TestMeta(t *testing.T) {
	type Entity struct {
		ID         int32     `gsd:"id,auto"`
		Name       string    `gsd:"name,pk,size:20"`
		Code       string    `gsd:"code"`
		CreateTime time.Time `gsd:"create_time,key:false,update:false,insert:true,select:true"`
		Ignore     bool      `gsd:"-"`
	}
	e := &Entity{
		ID:         100,
		Code:       "A100",
		CreateTime: time.Now(),
	}
	et := reflect.TypeOf(e)
	m := gsd.GetMeta(et)
	t.Log("ID: ", m.ID)

	// Properties
	assert.Equal(t, "entity", m.Table)
	assert.Equal(t, []string{"id", "name", "code", "create_time"}, m.Columns)
	assert.Equal(t, "id", m.Auto)
	assert.Equal(t, []string{"id", "name"}, m.PrimaryKeys)
	assert.Equal(t, []string{"name", "code", "create_time"}, m.Inserts)
	assert.Equal(t, []string{"code"}, m.Updates)
	assert.Equal(t, []string{"id", "name", "code", "create_time"}, m.Selects)

	// Value
	assert.Equal(t, e.ID, m.Value(e, "id"))

	// SetValue
	m.SetValue(e, "name", "xyz")
	assert.Equal(t, "xyz", e.Name)

	// Values
	values := m.Values(e, "id", "name")
	assert.Equal(t, []interface{}{e.ID, e.Name}, values)

	// Pointers
	ptrs := m.Pointers(e, "id", "name")
	assert.NotNil(t, ptrs)

	// SetAutoValue
	m.SetAutoValue(e, 200)
	assert.Equal(t, int32(200), e.ID)

	// CopyValues
	m.CopyValues(e, values, "id", "name")
	assert.Equal(t, []interface{}{int32(200), e.Name}, values)

	// InsertValues
	values = m.InsertValues(e) // name,code,create_time
	assert.Equal(t, []interface{}{e.Name, e.Code, e.CreateTime}, values)

	// CopyInsertValues
	values = make([]interface{}, 3)
	m.CopyInsertValues(e, values)
	assert.Equal(t, []interface{}{e.Name, e.Code, e.CreateTime}, values)

	// UpdateValues
	values = m.UpdateValues(e) // code
	assert.Equal(t, []interface{}{e.Code}, values)

	// CopyUpdateValues
	values = make([]interface{}, 1)
	m.CopyUpdateValues(e, values)
	assert.Equal(t, []interface{}{e.Code}, values)

	// KeyValues
	values = m.PrimaryKeyValues(e) // id,name
	assert.Equal(t, []interface{}{e.ID, e.Name}, values)

	// CopyKeyValues
	values = make([]interface{}, 2)
	m.CopyPrimaryKeyValues(e, values)
	assert.Equal(t, []interface{}{e.ID, e.Name}, values)
}

func BenchmarkMeta_Pointers(b *testing.B) {
	u := User{}
	ut := reflect.TypeOf(u)
	m := gsd.GetMeta(ut)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.Pointers(u, "id", "name")
	}
}

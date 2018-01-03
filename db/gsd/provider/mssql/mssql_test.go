package mssql

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/db/gsd"
	. "github.com/cuigh/auxo/db/gsd/abbr"
	"github.com/cuigh/auxo/test/assert"
)

const DBName = "mssql_test"

type User struct {
	ID         int32         `gsd:"id,pk"`
	Name       string        `gsd:"name"`
	Sex        bool          `gsd:"sex"`
	Age        gsd.NullInt32 `gsd:"age"`
	Salary     float32       `gsd:"salary"`
	CreateTime string        `gsd:"create_time,insert:false"`
}

func init() {
	config.AddFolder("../..")
	gsd.RegisterType(&User{}, &gsd.TableOptions{Name: "user"})
}

func TestOpen(t *testing.T) {
	db, err := gsd.Open("x")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestMustOpen(t *testing.T) {
	assert.Panic(t, func() {
		gsd.MustOpen("x")
	})
}

func TestDB_Create(t *testing.T) {
	db := gsd.MustOpen(DBName)
	user := &User{
		ID:   3,
		Name: "abc",
	}
	err := db.Create(user)
	assert.NoError(t, err)
}

func TestDB_CreateSlice(t *testing.T) {
	db := gsd.MustOpen(DBName)
	users := []*User{
		{ID: 3, Name: "abc"},
		{ID: 4, Name: "xyz"},
	}
	err := db.Create(users)
	assert.NoError(t, err)
}

func TestDB_Insert(t *testing.T) {
	db := gsd.MustOpen(DBName)
	r, err := db.Insert("user").Columns("id", "name").Values(1, "abc").Values(2, "xyz").Result()
	t.Log(r, err)
}

func TestDB_Remove(t *testing.T) {
	db := gsd.MustOpen(DBName)
	user := &User{
		ID: 3,
	}
	_, err := db.Remove(user)
	t.Log(err)
}

func TestDB_Delete(t *testing.T) {
	db := gsd.MustOpen(DBName)
	_, err := db.Delete("user").Where(Equal("id", 1)).Result()
	t.Log(err)
}

func TestDB_Update(t *testing.T) {
	db := gsd.MustOpen(DBName)
	_, err := db.Update("user").
		Set("name", "xyz").
		Inc("c1", 1).
		Dec("c2", 1).
		Expr("c3", "c4+10").
		Where(Equal("id", 1)).
		Result()
	t.Log(err)
}

func TestDB_Modify(t *testing.T) {
	db := gsd.MustOpen(DBName)
	user := &User{
		ID:   3,
		Name: "abc",
	}

	_, err := db.Modify(user)
	t.Log(err)

	_, err = db.Modify(user, Omit("code"))
	t.Log(err)
}

func TestDB_Load(t *testing.T) {
	db := gsd.MustOpen(DBName)

	user := &User{ID: 2}
	err := db.Load(user)
	assert.NoError(t, err)
	t.Log(user)

	user = &User{ID: -1}
	err = db.Load(user)
	assert.Same(t, gsd.ErrNoRows, err)
}

func TestDB_Select(t *testing.T) {
	db := gsd.MustOpen(DBName)

	// found
	user := &User{}
	err := db.Select("id", "name", "salary", "age", "sex", "create_time").
		From("user").
		Where(Equal("id", 2)).
		Fill(user)
	assert.NoError(t, err)
	t.Log(user)

	// missing
	err = db.Select("id", "name", "salary", "age", "sex", "create_time").
		From("user").
		Where(Equal("id", -1)).
		Fill(user)
	assert.Same(t, gsd.ErrNoRows, err)

	// full
	err = db.Query(C("id", "name", "salary", "age", "sex", "create_time"), true).
		From("user").
		Join("userinfo", On("id", "auto_id")).
		Where(Equal("id", -1)).
		GroupBy(C("age")).
		Having(Equal("age", 10)).
		OrderBy(C("id").ASC()).
		Limit(10, 10).
		Fill(user)
	assert.Same(t, gsd.ErrNoRows, err)
}

func BenchmarkDB_Create(b *testing.B) {
	user := &User{
		ID:   1,
		Name: "abc",
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen(DBName)
		err := db.Create(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDB_CreateSlice(b *testing.B) {
	users := []*User{
		{ID: 3, Name: "abc"},
		{ID: 4, Name: "xyz"},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen(DBName)
		err := db.Create(users)
		if err != nil {
			b.Fatal(err)
		}
	}
}

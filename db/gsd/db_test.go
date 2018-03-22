package gsd_test

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/db/gsd"
	. "github.com/cuigh/auxo/db/gsd/abbr"
	_ "github.com/cuigh/auxo/db/gsd/provider/mysql"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/test/assert"
)

type User struct {
	ID         int32         `gsd:"id,pk,auto"`
	Name       string        `gsd:"name"`
	Sex        bool          `gsd:"sex"`
	Age        gsd.NullInt32 `gsd:"age"`
	Salary     float32       `gsd:"salary"`
	CreateTime string        `gsd:"create_time,insert:false"`
}

func init() {
	config.AddFolder(".")
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
	db := gsd.MustOpen("test")
	user := &User{
		ID:   100,
		Name: "abc",
	}

	err := db.Create(user)
	assert.NoError(t, err)
	t.Log(user.ID)

	err = db.Create(user, Include("id", "name"))
	assert.NoError(t, err)
	t.Log(user.ID)
}

func TestDB_CreateSlice(t *testing.T) {
	db := gsd.MustOpen("test")
	users := []*User{
		{ID: 3, Name: "abc"},
		{ID: 4, Name: "xyz"},
	}
	err := db.Create(users)
	assert.NoError(t, err)
}

func TestDB_Insert(t *testing.T) {
	db := gsd.MustOpen("test")
	r, err := db.Insert("user").Columns("id", "name").Values(1, "abc").Values(2, "xyz").Result()
	t.Log(r, err)
}

func TestDB_Remove(t *testing.T) {
	db := gsd.MustOpen("test")
	user := &User{
		ID: 3,
	}
	_, err := db.Remove(user)
	t.Log(err)
}

func TestDB_Delete(t *testing.T) {
	db := gsd.MustOpen("test")
	_, err := db.Delete("user").Where(Equal("id", 1)).Result()
	t.Log(err)
}

func TestDB_Update(t *testing.T) {
	db := gsd.MustOpen("test")
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
	db := gsd.MustOpen("test")
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
	db := gsd.MustOpen("test")

	user := &User{ID: 2}
	err := db.Load(user)
	assert.NoError(t, err)
	t.Log(user)

	user = &User{ID: -1}
	err = db.Load(user)
	assert.Same(t, gsd.ErrNoRows, err)
}

func TestDB_Select(t *testing.T) {
	db := gsd.MustOpen("test")

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

func TestDB_Select_List(t *testing.T) {
	db := gsd.MustOpen("test")

	users := make([]*User, 0)
	var count int
	err := db.Select("id", "name", "salary", "age", "sex", "create_time").
		From("user").
		List(&users, &count)
	assert.NoError(t, err)
	t.Logf("count: %v", count)
}

func TestDB_Select_Fill(t *testing.T) {
	db := gsd.MustOpen("test")

	users := make([]*User, 0)
	err := db.Select("id", "name", "salary", "age", "sex", "create_time").
		From("user").
		Fill(&users)
	assert.NoError(t, err)
	t.Logf("count: %v", len(users))
}

func TestDB_Count(t *testing.T) {
	db := gsd.MustOpen("test")

	var count int
	err := db.Count("user").Scan(&count)
	assert.NoError(t, err)
	t.Log("count: ", count)

	count, err = db.Count("user").Value()
	t.Log(count, err)
}

func TestDB_Execute(t *testing.T) {
	var (
		db   = gsd.MustOpen("test")
		err  error
		id   int32
		name string
	)

	// Value
	name, err = db.Execute("select name from user where id = ?", 1).Value().String()
	assert.NoError(t, err)
	t.Log("name:", name)

	// Scan
	err = db.Execute("select id, name from user where id = ?", 1).Scan(&id, &name)
	assert.NoError(t, err)

	// Result
	r, err := db.Execute("delete from user where id = ?", -1).Result()
	assert.NoError(t, err)
	t.Log(r.RowsAffected())

	// Reader
	reader, err := db.Execute("select id, name from user").Reader()
	assert.NoError(t, err)
	defer reader.Close()
	for reader.Next() {
		reader.Scan(&id, &name)
		t.Log(id, name)
	}
	assert.NoError(t, reader.Err())

	// For
	err = db.Execute("select id, name from user").For(func(reader gsd.Reader) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestDB_Transact(t *testing.T) {
	db := gsd.MustOpen("test")

	// Commit
	err := db.Transact(func(tx gsd.TX) error {
		user := &User{ID: 1}
		tx.Load(user)
		return nil
	})
	assert.NoError(t, err)

	// Cancel
	err = db.Transact(func(tx gsd.TX) error {
		user := &User{
			ID:   1,
			Name: "abc",
		}
		tx.Create(user)
		return gsd.ErrTXCancelled
	})
	assert.NoError(t, err)

	// Panic
	err = db.Transact(func(tx gsd.TX) error {
		panic(errors.New("test TX panic"))
	})
	assert.Error(t, err)
}

func BenchmarkDB_Create(b *testing.B) {
	user := &User{
		ID:   1,
		Name: "abc",
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen("test")
		db.Create(user)
	}
}

func BenchmarkDB_CreateSlice(b *testing.B) {
	users := []*User{
		{ID: 3, Name: "abc"},
		{ID: 4, Name: "xyz"},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen("test")
		db.Create(users)
	}
}

func BenchmarkDB_Load(b *testing.B) {
	user := &User{
		ID: 1,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen("test")
		if err := db.Load(user); err != nil {
			b.Fatal(err)
		}
	}
}

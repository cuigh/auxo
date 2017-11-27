package mysql

import (
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/test/assert"
)

type User struct {
	ID   int32  `gsd:"id,pk"`
	Name string `gsd:"user_name"`
}

func init() {
	config.AddFolder("../..")
	gsd.RegisterType(&User{}, &gsd.TableOptions{Name: "user"})
}

func TestCreate(t *testing.T) {
	db := gsd.MustOpen("test")

	user := &User{3, "abc"}
	err := db.Create(user)
	assert.NoError(t, err)
}

func TestCreateSlice(t *testing.T) {
	db := gsd.MustOpen("test")

	users := []*User{
		{3, "abc"},
		{4, "xyz"},
	}
	err := db.Create(users)
	assert.NoError(t, err)
}

func BenchmarkCreate(b *testing.B) {
	user := &User{1, "abc"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen("test")
		err := db.Create(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateSlice(b *testing.B) {
	users := []*User{
		{3, "abc"},
		{4, "xyz"},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db := gsd.MustOpen("test")
		err := db.Create(users)
		if err != nil {
			b.Fatal(err)
		}
	}
}

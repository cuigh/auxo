package proxy_test

import (
	"testing"

	"github.com/cuigh/auxo/config"
	_ "github.com/cuigh/auxo/db/gsd/provider/mysql"
	"github.com/cuigh/auxo/db/gsd/proxy"
	"github.com/cuigh/auxo/test/assert"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	config.AddFolder("..")
}

func TestProxy(t *testing.T) {
	type User struct {
		ID   int32 `gsd:",auto"`
		Name string
	}
	type UserDao struct {
		Get    func(id int32) (*User, error) `gsd:"find,cache:user.load"`
		Load   func(u *User) error           `gsd:"load,cache:user.load"`
		Remove func(u *User) error           `gsd:"remove"`
		//Delete func(id int32) error          `gsd:"remove,table:user"`
		Create func(u *User) error `gsd:"create"`
		Update func(u *User) error `gsd:"modify"`
		//Search func(f *UserFilter) ([]*User, int, error) `gsd:"search,cache:user.search"`
	}

	dao := &UserDao{}
	proxy.Apply("test", dao)

	user, err := dao.Get(1)
	assert.NoError(t, err)
	t.Log(user)

	err = dao.Load(user)
	assert.NoError(t, err)
	t.Log(user)

	user = &User{ID: 2, Name: "noname"}
	err = dao.Create(user)
	assert.NoError(t, err)

	user.Name = "noname111"
	err = dao.Update(user)
	assert.NoError(t, err)

	err = dao.Remove(user)
	assert.NoError(t, err)
}

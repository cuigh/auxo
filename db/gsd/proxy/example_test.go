package proxy_test

import (
	"github.com/cuigh/auxo/db/gsd/proxy"
)

func ExampleProxy_Apply() {
	type User struct {
		ID   int32 `gsd:",auto"`
		Name string
	}
	type UserDao struct {
		Get    func(id int32) (*User, error) `gsd:"find,cache:user.load"`
		Load   func(u *User) error           `gsd:"load,cache:user.load"`
		Remove func(u *User) error           `gsd:"remove"`
		Create func(u *User) error           `gsd:"create"`
		Update func(u *User) error           `gsd:"modify"`
	}

	dao := &UserDao{}
	proxy.Apply("test", dao)

	user, err := dao.Get(1)
	if err != nil {
		panic(err)
	}

	err = dao.Load(user)
	if err != nil {
		panic(err)
	}

	user = &User{ID: 2, Name: "noname"}
	err = dao.Create(user)
	if err != nil {
		panic(err)
	}

	user.Name = "noname"
	err = dao.Update(user)
	if err != nil {
		panic(err)
	}

	err = dao.Remove(user)
	if err != nil {
		panic(err)
	}
}

# proxy

**proxy** implements a dynamic proxy for gsd, like **JPA** in Java.

## Usage

With **proxy**, you only need add interface and tags, **proxy** create these methods for you automatically.

```go
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

func ProxyTest() {
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
```

## Create custom proxy method

**proxy** support 5 proxy methods now, you can create your own proxy method easily.

```go
proxy.Register("search", func(db *gsd.LazyDB, ft reflect.Type, options data.Options) reflect.Value {
    // todo: impliment proxy method
})
```

Once you register it, you can use it in any dao struct.

```go
type UserDao struct {
	Search func(name string) ([]*User, error) `gsd:"search"`
}
```
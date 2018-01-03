# gsd

gsd is a lightweight, fluent SQL data access library. It supports various types of database, like mysql/mssql/sqlite etc.

* High performance
* Support context.Context
* Support opentracing

## Configure

**gsd** use **auxo/config** package to manage database options. There is a sample config file in the package(app.yml):

```yaml
db.sql:
    test:
      provider: mysql
      address: localhost:27017/test
      trace:
        enabled: true
        time: 1s
      options:
        max_open_conns: 100
        max_idle_conns: 1
```

Once you add configuration to app.yml, you can open a database like this:

```
db, err := gsd.Open("Test")
......
```

## Usage

The API of **gsd** is very similar to native SQL expression.

Let's define a dummy User struct first.

```go
type User struct {
	ID         int32         `gsd:"id,pk,auto"`
	Name       string        `gsd:"name"`
	Sex        bool          `gsd:"sex"`
	Age        gsd.NullInt32 `gsd:"age"`
	Salary     float32       `gsd:"salary"`
	CreateTime string        `gsd:"create_time,insert:false"`
}
```

### INSERT

```go
r, err := db.Insert("user").Columns("id", "name").Values(1, "abc").Values(2, "xyz").Result()
// ...
```

OR

```go
user := &User{
	ID:   100,
	Name: "abc",
}
err := db.Create(user)
// or specify columns
// err = db.Create(user, Include("id", "name"))
```

### DELETE

```go
r, err := db.Delete("user").Where(Equal("id", 1)).Result()
// ...
```

OR

```go
user := &User{
	ID: 3,
}
r, err := db.Remove(user)
```

### UPDATE

```go
r, err := db.Update("user").
		Set("name", "xyz").
		Inc("c1", 1).
		Dec("c2", 1).
		Expr("c3", "c4+10").
		Where(Equal("id", 1)).
		Result()
```

OR

```go
r, err := db.Modify(user)
// or specify columns
r, err = db.Modify(user, Omit("code"))
```

### SELECT

```go
err := db.Select("id", "name", "salary", "age", "sex", "create_time").
	From("user").
	Where(Equal("id", 2)).
	Fill(user)
```

A more complex example

```go
err = db.Query(C("id", "name", "salary", "age", "sex", "create_time"), true).
	From("user").
	Join("userinfo", On("id", "auto_id")).
	Where(Equal("id", -1)).
	GroupBy(C("age")).
	Having(Equal("age", 10)).
	OrderBy(C("id").ASC()).
	Limit(10, 10).
	Fill(user)
```

OR

```go
user := &User{
	ID: 3,
}
err := db.Load(user)
```

### COUNT

```go
err := db.Count("user").Scan(&count)
// or
count, err = db.Count("user").Value()
```

### TRANSACTION

```go
err := db.Transact(func(tx gsd.TX) error {
	count, err := tx.Count("user").Value()
	t.Log(count, err)
	return err
})
if err != nil {
	log.Fatal(err)
}
```

## TODO

* More proxy actions.

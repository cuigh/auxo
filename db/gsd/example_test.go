package gsd_test

import (
	"context"
	"fmt"

	"github.com/cuigh/auxo/db/gsd"
	. "github.com/cuigh/auxo/db/gsd/abbr"
)

func ExampleDB_Load() {
	db := gsd.MustOpen("test")

	user := &User{ID: 1}
	err := db.Load(context.TODO(), user)
	if err != nil {
		fmt.Println("Load failed:", err)
	}
	fmt.Println("Name:", user.Name)

	user = &User{ID: -1}
	err = db.Load(context.TODO(), user)
	if err == gsd.ErrNoRows {
		fmt.Println("Error:", err.Error())
	}
	// Output:
	// Name: abc
	// Error: sql: no rows in result set
}

func ExampleDB_Select() {
	db := gsd.MustOpen("test")

	user := &User{}
	err := db.Select(context.TODO(), "id", "name", "salary", "age", "sex", "create_time").
		From("user").
		Where(Equal("id", 1)).
		Fill(user)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	fmt.Println("Name:", user.Name)
	// Output:
	// Name: abc
}

func ExampleDB_Remove() {
	db := gsd.MustOpen("test")
	user := &User{
		ID: 3,
	}
	_, err := db.Remove(context.TODO(), user)
	fmt.Println("Error:", err)
	// Output:
	// Error: <nil>
}

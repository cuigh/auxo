package gsd_test

import (
	"context"
	"testing"

	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/test/assert"
)

func TestStmt(t *testing.T) {
	db := gsd.MustOpen("test")

	stmt, err := db.Prepare(context.TODO(), "select id, name from user where id = ?")
	assert.NoError(t, err)

	user := &User{}
	err = stmt.Execute(1).Fill(user)
	assert.NoError(t, err)
}

func BenchmarkStmt_Execute(b *testing.B) {
	db := gsd.MustOpen("test")

	stmt, err := db.Prepare(context.TODO(), "select id, name from user where id = ?")
	assert.NoError(b, err)

	user := &User{}
	stmt.Execute(1).Fill(user)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user := &User{}
		err = stmt.Execute(1).Fill(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

package gsd_test

import (
	"context"
	"testing"

	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/test/assert"
)

func TestTX_Commit(t *testing.T) {
	db := gsd.MustOpen("test")
	err := db.Transact(context.TODO(), func(tx gsd.TX) error {
		count, err := tx.Count("user").Value()
		t.Log(count, err)
		return err
	})
	assert.NoError(t, err)
}

// todo:
func TestTX_Rollback(t *testing.T) {
	db := gsd.MustOpen("test")
	err := db.Transact(context.TODO(), func(tx gsd.TX) error {
		count, err := tx.Count("user").Value()
		t.Log(count, err)
		return err
	})
	assert.NoError(t, err)
}

// todo:
func TestTX_Cancel(t *testing.T) {
	db := gsd.MustOpen("test")
	err := db.Transact(context.TODO(), func(tx gsd.TX) error {
		count, err := tx.Count("user").Value()
		t.Log(count, err)
		return err
	})
	assert.NoError(t, err)
}

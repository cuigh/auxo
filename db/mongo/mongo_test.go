package mongo

import (
	"context"
	"testing"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/test/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	config.AddFolder(".")
}

func TestOpen(t *testing.T) {
	db := MustOpen("test")
	_, err := db.Collection("user").CountDocuments(context.TODO(), bson.M{})
	assert.NoError(t, err)
}

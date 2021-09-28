package mongo

import (
	"context"
	"time"

	"github.com/cuigh/auxo/app"
	"github.com/cuigh/auxo/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func MustOpen(name string) *mongo.Database {
	db, err := Open(name)
	if err != nil {
		panic(err)
	}
	return db
}

func Open(name string) (*mongo.Database, error) {
	db := name
	opts := &options.ClientOptions{}

	err := config.UnmarshalOption("db.mongo."+name, opts)
	if err != nil {
		return nil, err
	}

	if addr := config.GetString("db.mongo." + name + ".address"); addr != "" {
		opts.ApplyURI(addr)
		if cs, err := connstring.Parse(addr); err == nil && cs.Database != "" {
			db = cs.Database
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts.SetAppName(app.Name))
	if err != nil {
		panic(err)
	}

	return client.Database(db), nil
}

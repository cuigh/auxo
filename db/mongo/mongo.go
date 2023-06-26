package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
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

type Table[T, K any] struct {
	*mongo.Collection
}

func NewTable[T, K any](db *mongo.Database, name string, opts ...*options.CollectionOptions) *Table[T, K] {
	return &Table[T, K]{db.Collection(name, opts...)}
}

func (t *Table[T, K]) QueryByFilter(ctx context.Context, filter any) (r *T, err error) {
	var v T
	if err = t.FindOne(ctx, filter).Decode(&v); err == nil {
		r = &v
	} else if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}

func (t *Table[T, K]) Query(ctx context.Context, id K) (r *T, err error) {
	return t.QueryByFilter(ctx, bson.M{"_id": id})
}

// Create insert a new document to database.
func (t *Table[T, K]) Create(ctx context.Context, doc *T) (err error) {
	_, err = t.InsertOne(ctx, doc)
	return
}

// Update updates document by ID.
func (t *Table[T, K]) Update(ctx context.Context, id K, update any) (err error) {
	_, err = t.UpdateByID(ctx, id, update)
	return
}

// Upsert updates document by ID if found, otherwise insert document.
func (t *Table[T, K]) Upsert(ctx context.Context, id K, update any) (err error) {
	_, err = t.UpdateByID(ctx, id, update, options.Update().SetUpsert(true))
	return
}

// Delete deletes document by ID.
func (t *Table[T, K]) Delete(ctx context.Context, id K) (err error) {
	_, err = t.DeleteOne(ctx, bson.M{"_id": id})
	return
}

// FetchByPage returns qualified records with paging.
func (t *Table[T, K]) FetchByPage(ctx context.Context, pageIndex, pageSize int64, filter, sorter any) (records []*T, count int64, err error) {
	// fetch total count
	count, err = t.CountDocuments(ctx, filter)
	if err != nil {
		return
	}

	// fetch records
	var (
		cur  *mongo.Cursor
		opts = options.Find().SetSkip(pageSize * (pageIndex - 1)).SetLimit(pageSize)
	)
	if filter == nil {
		filter = bson.M{}
	}
	if sorter != nil {
		opts.SetSort(sorter)
	}
	cur, err = t.Collection.Find(ctx, filter, opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &records)
	return
}

// Fetch returns all qualified records.
func (t *Table[T, K]) Fetch(ctx context.Context, filter, sorter any) (records []*T, err error) {
	var cur *mongo.Cursor
	if filter == nil {
		filter = bson.M{}
	}
	if sorter == nil {
		cur, err = t.Collection.Find(ctx, filter)
	} else {
		cur, err = t.Collection.Find(ctx, filter, options.Find().SetSort(sorter))
	}
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &records)
	return
}

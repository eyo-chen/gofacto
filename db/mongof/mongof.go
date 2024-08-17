package mongof

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eyo-chen/gofacto/internal/db"
)

// config is for MongoDB configuration
type config struct {
	// db is the database connection
	db *mongo.Database
}

// NewConfig creates a new MongoDB configuration
func NewConfig(db *mongo.Database) *config {
	return &config{
		db: db,
	}
}

func (c *config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	res, err := c.db.Collection(params.StorageName).InsertOne(ctx, params.Value)
	if err != nil {
		return nil, err
	}

	id := res.InsertedID.(primitive.ObjectID)
	setIDField(params.Value, id)
	return params.Value, nil
}

func (c *config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	res, err := c.db.Collection(params.StorageName).InsertMany(ctx, params.Values)
	if err != nil {
		return nil, err
	}

	for i, rawID := range res.InsertedIDs {
		id := rawID.(primitive.ObjectID)
		setIDField(params.Values[i], id)
	}

	return params.Values, nil
}

func (c *config) GenCustomType(t reflect.Type) (interface{}, bool) {
	return nil, false
}

// setIDField sets the ID field of the value to the given ID
func setIDField(val interface{}, id primitive.ObjectID) {
	v := reflect.ValueOf(val).Elem().FieldByName("ID")
	if v.IsValid() && v.CanSet() {
		v.Set(reflect.ValueOf(id))
	}
}

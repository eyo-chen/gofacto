package gofacto

import (
	"context"
	"reflect"

	"github.com/eyo-chen/gofacto/internal/db"
)

// database is responsible for inserting data into the database
type database interface {
	// insert inserts a single data into the database
	Insert(context.Context, db.InsertParams) (interface{}, error)

	// insertList inserts a list of data into the database
	InsertList(context.Context, db.InsertListParams) ([]interface{}, error)

	// GenCustomType generates a non-zero value for custom types
	GenCustomType(reflect.Type) (interface{}, bool)
}

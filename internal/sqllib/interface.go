package sqllib

import (
	"context"

	"github.com/eyo-chen/gofacto/db"
)

// SQLHandler is the interface for raw SQL database operations
type SQLHandler interface {
	// Insert inserts a single data into the database
	Insert(ctx context.Context, params db.InserParams) (interface{}, error)

	// InsertList inserts a list of data into the database
	InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error)
}

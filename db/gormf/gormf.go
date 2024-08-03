package gormf

import (
	"context"
	"reflect"
	"time"

	"github.com/eyo-chen/gofacto/db"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// config is for Gorm configuration
type config struct {
	// db is the database connection
	db *gorm.DB
}

// NewConfig creates a new gorm configuration
func NewConfig(db *gorm.DB) *config {
	return &config{
		db: db,
	}
}

func (c *config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	if err := c.db.WithContext(ctx).Table(params.StorageName).Create(params.Value).Error; err != nil {
		return nil, err
	}

	return params.Value, nil
}

func (c *config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	// NOTE: Using for-loop to insert is a workaround for GORM
	for _, v := range params.Values {
		if err := c.db.WithContext(ctx).Table(params.StorageName).Create(v).Error; err != nil {
			return nil, err
		}
	}

	return params.Values, nil
}

func (c *config) GenCustomType(t reflect.Type) (interface{}, bool) {
	// Check if the type is a pointer
	if t.Kind() == reflect.Ptr {
		v, ok := c.GenCustomType(t.Elem())
		if !ok {
			return nil, false
		}

		ptr := reflect.New(reflect.TypeOf(v))
		ptr.Elem().Set(reflect.ValueOf(v))
		return ptr.Interface(), true
	}

	// Handle specific types
	switch t.String() {
	case jsonType:
		return datatypes.JSON([]byte(`{"test": "test"}`)), true
	case dateType:
		return datatypes.Date(time.Now()), true
	case timeType:
		return datatypes.NewTime(1, 2, 3, 0), true
	default:
		return nil, false
	}
}

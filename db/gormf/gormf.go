package gormf

import (
	"context"

	"github.com/eyo-chen/gofacto/db"
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

package gormf

import (
	"context"

	"github.com/eyo-chen/gofacto/db"
	"gorm.io/gorm"
)

// Config is for Gorm database operations
type Config struct {
	// DB is the database connection
	// must provide if want to insert data into the database
	DB *gorm.DB
}

func (s *Config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	if err := s.DB.WithContext(ctx).Table(params.StorageName).Create(params.Value).Error; err != nil {
		return nil, err
	}

	return params.Value, nil
}

func (s *Config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	// NOTE: Using for-loop to insert is a workaround for GORM
	for _, v := range params.Values {
		if err := s.DB.WithContext(ctx).Table(params.StorageName).Create(v).Error; err != nil {
			return nil, err
		}
	}

	return params.Values, nil
}

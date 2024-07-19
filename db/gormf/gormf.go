package gormf

import (
	"context"
	"fmt"

	"github.com/eyo-chen/gofacto/db"
	"gorm.io/gorm"
)

// Config is for Gorm database operations
type Config struct {
	// DB is the database connection
	// must provide if want to insert data into the database
	DB *gorm.DB

	// Ctx is the context for the database operations
	// it is optional
	Ctx context.Context
}

func (s *Config) Insert(params db.InserParams) (interface{}, error) {
	if s.Ctx != nil {
		if err := s.DB.WithContext(s.Ctx).Table(params.StorageName).Create(params.Value).Error; err != nil {
			return nil, err
		}

		return params.Value, nil
	}

	if err := s.DB.Table(params.StorageName).Create(params.Value).Error; err != nil {
		return nil, err
	}

	return params.Value, nil
}

func (s *Config) InsertList(params db.InserListParams) ([]interface{}, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("database connection is not provided")
	}

	// NOTE: Using for-loop to insert is a workaround for GORM
	if s.Ctx != nil {
		for _, v := range params.Values {
			if err := s.DB.WithContext(s.Ctx).Table(params.StorageName).Create(v).Error; err != nil {
				return nil, err
			}
		}

		return params.Values, nil
	}

	for _, v := range params.Values {
		if err := s.DB.Table(params.StorageName).Create(v).Error; err != nil {
			return nil, err
		}
	}

	return params.Values, nil
}

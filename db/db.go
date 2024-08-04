package db

import (
	"context"
	"reflect"
)

// Database is responsible for inserting data into the database
type Database interface {
	// insert inserts a single data into the database
	Insert(context.Context, InserParams) (interface{}, error)

	// insertList inserts a list of data into the database
	InsertList(context.Context, InserListParams) ([]interface{}, error)

	// GenCustomType generates a non-zero value for custom types
	GenCustomType(reflect.Type) (interface{}, bool)
}

// InsertParams is a struct that holds the parameters for the Insert method
type InserParams struct {
	StorageName string
	Value       interface{}
}

// InserListParams is a struct that holds the parameters for the InsertList method
type InserListParams struct {
	StorageName string
	Values      []interface{}
}

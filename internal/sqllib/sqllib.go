package sqllib

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/eyo-chen/gofacto/db"
	"github.com/eyo-chen/gofacto/internal/utils"
)

// Config is for raw SQL database operations
type Config struct {
	// db is the database connection
	db *sql.DB

	// dialect is the SQL dialect
	dialect sqlDialect

	// packageName is the package name
	packageName string
}

// sqlDialect defines the behavior for different SQL dialects
type sqlDialect interface {
	// GenPlaceholder generates a placeholder
	GenPlaceholder(placeholderIdx int) string

	// GenInsertStmt generates an insert raw SQL statement
	GenInsertStmt(tableName, fieldNames, placeholder string) string

	// InsertToDB inserts the values to the database
	InsertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error)
}

// NewConfig initializes a sqllib config for raw SQL database operations
func NewConfig(db *sql.DB, dialect sqlDialect, packageName string) db.Database {
	return &Config{
		db:          db,
		dialect:     dialect,
		packageName: packageName,
	}
}

func (c *Config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	rawStmt, vals := c.prepareStmtAndVals(params.StorageName, params.Value)

	stmt, err := c.db.Prepare(rawStmt)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) && err == nil {
			err = rollbackErr
		}
	}()

	id, err := c.dialect.InsertToDB(ctx, tx, stmt, vals[0])
	if err != nil {
		return nil, err
	}

	setIDField(params.Value, id)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return params.Value, nil
}

func (c *Config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	rawStmt, fieldValues := c.prepareStmtAndVals(params.StorageName, params.Values...)

	stmt, err := c.db.Prepare(rawStmt)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) && err == nil {
			err = rollbackErr
		}
	}()

	result := make([]interface{}, len(fieldValues))
	for i, vals := range fieldValues {
		id, err := c.dialect.InsertToDB(ctx, tx, stmt, vals)
		if err != nil {
			return nil, err
		}

		v := params.Values[i]
		setIDField(v, id)

		result[i] = v
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Config) GenCustomType(t reflect.Type) (interface{}, bool) {
	return nil, false
}

// prepareStmtAndVals prepares the SQL insert statement and the values to be inserted
// values are the pointer to the struct
func (c *Config) prepareStmtAndVals(tableName string, values ...interface{}) (string, [][]interface{}) {
	fieldNames := []string{}
	placeholders := []string{}
	fieldValues := [][]interface{}{}

	for index, val := range values {
		val := reflect.ValueOf(val).Elem()
		vals := []interface{}{}

		placeholderIndex := 1
		for i := 0; i < val.NumField(); i++ {
			n := val.Type().Field(i).Name
			if n == "ID" {
				continue
			}

			vals = append(vals, val.Field(i).Interface())

			if index == 0 {
				fieldName := val.Type().Field(i).Tag.Get(c.packageName)
				if fieldName == "" {
					fieldName = utils.CamelToSnake(n)
				}

				fieldNames = append(fieldNames, fieldName)
				placeholders = append(placeholders, c.dialect.GenPlaceholder(placeholderIndex))
			}

			placeholderIndex++
		}

		fieldValues = append(fieldValues, vals)
	}

	// construct the SQL insert statement
	fns := strings.Join(fieldNames, ", ")
	phs := strings.Join(placeholders, ", ")
	rawStmt := c.dialect.GenInsertStmt(tableName, fns, phs)

	return rawStmt, fieldValues
}

// setIDField sets the id value on ID field of the given value
func setIDField(v interface{}, id int64) {
	val := reflect.ValueOf(v).Elem()
	idField := val.FieldByName("ID")
	if !idField.IsValid() || !idField.CanSet() {
		return
	}

	switch idField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		idField.SetInt(id)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		idField.SetUint(uint64(id))
	}
}

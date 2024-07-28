package mysqlf

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/eyo-chen/gofacto/db"
	"github.com/eyo-chen/gofacto/internal/utils"
)

// config is for MySQL database configuration
type config struct {
	// db is the database connection
	db *sql.DB
}

// NewConfig creates a new MySQL configuration
func NewConfig(db *sql.DB) *config {
	return &config{
		db: db,
	}
}

func (c *config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	rawStmt, vals := prepareStmtAndVals(params.StorageName, params.Value)

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

	id, err := insertToDB(ctx, tx, stmt, vals[0])
	if err != nil {
		return nil, err
	}

	setIDField(params.Value, id)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return params.Value, nil
}

func (c *config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	rawStmt, fieldValues := prepareStmtAndVals(params.StorageName, params.Values...)

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
		id, err := insertToDB(ctx, tx, stmt, vals)
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

func (c *config) SetIDField(v interface{}, i int) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("SetIDField: argument must be a pointer")
	}

	if reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("SetIDField: argument must be a pointer to a struct")
	}

	val := reflect.ValueOf(v).Elem()
	idField := val.FieldByName("ID")

	if !idField.IsValid() {
		return fmt.Errorf("SetIDField: ID field not found")
	}

	if !idField.CanSet() {
		return fmt.Errorf("SetIDField: ID field is not settable")
	}

	switch idField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		idField.SetInt(int64(i))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		idField.SetUint(uint64(i))
	}

	return nil
}

// prepareStmtAndVals prepares the SQL insert statement and the values to be inserted
// values are the pointer to the struct
func prepareStmtAndVals(tableName string, values ...interface{}) (string, [][]interface{}) {
	fieldNames := []string{}
	placeholders := []string{}
	fieldValues := [][]interface{}{}

	for index, val := range values {
		val := reflect.ValueOf(val).Elem()
		vals := []interface{}{}

		for i := 0; i < val.NumField(); i++ {
			n := val.Type().Field(i).Name
			if n == "ID" {
				continue
			}

			vals = append(vals, val.Field(i).Interface())

			if index == 0 {
				fieldName := val.Type().Field(i).Tag.Get("sqlf")
				if fieldName == "" {
					fieldName = utils.CamelToSnake(n)
				}

				fieldNames = append(fieldNames, fieldName)
				placeholders = append(placeholders, "?")
			}
		}

		fieldValues = append(fieldValues, vals)
	}

	// construct the SQL insert statement
	fns := strings.Join(fieldNames, ", ")
	phs := strings.Join(placeholders, ", ")
	rawStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, fns, phs)

	return rawStmt, fieldValues
}

// insertToDB inserts the given values to the database
func insertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error) {
	res, err := tx.Stmt(stmt).ExecContext(ctx, vals...)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// setIDField sets the id value on ID field of the given value
func setIDField(v interface{}, id int64) {
	val := reflect.ValueOf(v).Elem()
	idField := val.FieldByName("ID")
	switch idField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		idField.SetInt(id)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		idField.SetUint(uint64(id))
	}
}

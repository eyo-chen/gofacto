package sqlf

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/eyo-chen/gofacto/db"
)

// Config is for raw SQL database operations
type Config struct {
	// DB is the database connection
	// must provide if want to insert data into the database
	DB *sql.DB
}

func (c *Config) Insert(ctx context.Context, params db.InserParams) (interface{}, error) {
	rawStmt, vals := prepareStmtAndVals(params.StorageName, params.Value)

	// Prepare the insert statement
	stmt, err := c.DB.Prepare(rawStmt)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := c.DB.Begin()
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

func (c *Config) InsertList(ctx context.Context, params db.InserListParams) ([]interface{}, error) {
	rawStmt, fieldValues := prepareStmtAndVals(params.StorageName, params.Values...)

	stmt, err := c.DB.Prepare(rawStmt)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	tx, err := c.DB.Begin()
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

func (c *Config) SetIDField(v interface{}, i int) error {
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
					fieldName = camelToSnake(n)
				}

				fieldNames = append(fieldNames, fieldName)
				placeholders = append(placeholders, "?")
			}
		}

		fieldValues = append(fieldValues, vals)
	}

	// Construct the SQL insert statement
	fns := strings.Join(fieldNames, ", ")
	phs := strings.Join(placeholders, ", ")
	rawStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, fns, phs)

	return rawStmt, fieldValues
}

// insertToDB inserts the given values to the database
func insertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error) {
	var res sql.Result
	var errSQL error

	res, errSQL = tx.Stmt(stmt).ExecContext(ctx, vals...)
	if errSQL != nil {
		return 0, errSQL
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

// CamelToSnake converts the given camel case string to snake case
func camelToSnake(input string) string {
	var buf bytes.Buffer

	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(input[i-1])) {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

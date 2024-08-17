package mysqlf

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eyo-chen/gofacto/internal/sqllib"
)

// NewConfig initializes interface for raw mySQL database operations
func NewConfig(db *sql.DB) *sqllib.Config {
	return sqllib.NewConfig(db, &mySQLDialect{}, "mysqlf")
}

// mySQLDialect defines the behavior for MySQL SQL dialect
type mySQLDialect struct{}

func (d *mySQLDialect) GenPlaceholder(_ int) string {
	return "?"
}

func (d *mySQLDialect) GenInsertStmt(tableName, fieldNames, placeholder string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, fieldNames, placeholder)
}

func (d *mySQLDialect) InsertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error) {
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

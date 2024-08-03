package postgresf

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eyo-chen/gofacto/db"
	"github.com/eyo-chen/gofacto/internal/sqllib"
)

// NewConfig initializes interface for raw PostgreSQL database operations
func NewConfig(db *sql.DB) db.Database {
	return sqllib.NewConfig(db, &postgresDialect{}, "postgresf")
}

// postgresDialect defines the behavior for PostgreSQL SQL dialect
type postgresDialect struct{}

func (d *postgresDialect) GenPlaceholder(placeholderIndex int) string {
	return fmt.Sprintf("$%d", placeholderIndex)
}

func (d *postgresDialect) GenInsertStmt(tableName, fieldNames, placeholder string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, fieldNames, placeholder)
}

func (d *postgresDialect) InsertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error) {
	var id int64
	err := tx.Stmt(stmt).QueryRowContext(ctx, vals...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

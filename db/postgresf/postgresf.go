package postgresf

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eyo-chen/gofacto/internal/sqllib"
)

// Config is for raw SQL database operations
type config struct {
	sqllib.Config
}

func NewConfig(db *sql.DB) *config {
	return &config{
		Config: sqllib.Config{
			DB:          db,
			Dialect:     &mySQLDialect{},
			PackageName: "postgresf",
		},
	}
}

type mySQLDialect struct{}

func (d *mySQLDialect) GenPlaceholder(placeholderIndex int) string {
	return fmt.Sprintf("$%d", placeholderIndex)
}

func (d *mySQLDialect) GenInsertStmt(tableName, fieldNames, placeholder string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, fieldNames, placeholder)

}

func (d *mySQLDialect) InsertToDB(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, vals []interface{}) (int64, error) {
	var id int64
	err := tx.Stmt(stmt).QueryRowContext(ctx, vals...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

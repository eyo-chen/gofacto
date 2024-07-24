package mysqlf

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
			PackageName: "mysqlf",
		},
	}
}

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

package gormf

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/eyo-chen/gofacto/db/docker"
	_ "github.com/go-sql-driver/mysql"
)

func TestInsert(t *testing.T) {
	port := docker.RunDocker(docker.ImageMySQL)
	dba, err := sql.Open("mysql", fmt.Sprintf("root:root@(localhost:%s)/mysql?parseTime=true", port))
	if err != nil {
		log.Fatalf("sql.Open failed: %s", err)
	}

	defer func() {
		dba.Close()
		docker.PurgeDocker()
	}()

	// Read SQL file
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %s", err)
	}

	// Split SQL file content into individual statements
	queries := strings.Split(string(schema), ";")

	// Execute SQL statements one by one
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		fmt.Println("Executing query:", query)
		if _, err := dba.Exec(query); err != nil {
			log.Fatalf("Failed to execute query: %s, error: %s", query, err)
		}
	}

	insertStmt := "INSERT INTO users (username) VALUES (?)"
	if _, err := dba.Exec(insertStmt, "test"); err != nil {
		log.Fatalf("Failed to insert: %s", err)
	}

	type users struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	selectStmt := "SELECT id, username FROM users"
	rows, err := dba.Query(selectStmt)
	if err != nil {
		log.Fatalf("Failed to query: %s", err)
	}

	defer rows.Close()

	for rows.Next() {
		var u users
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			log.Fatalf("Failed to scan: %s", err)
		}

		fmt.Println("user", u)
	}

	// Additional test logic can go here

	fmt.Println("db", dba)
}

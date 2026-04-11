package storage

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "postgres://user:pass@localhost:5432/myapp_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	return db
}

func Test_CreateMigrationTable(t *testing.T) {
	db := setupTestDB()

	var exists bool
	tx.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.columns
		WHERE table_name = 'migration' AND column_name = 'version'
	)`).Scan(&exists)

	if !exists {
		t.Error("expected table migration with version column to exist")
	}
}

package storage

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// TODO: read password from env
	dsn := "postgres://postgres:password@localhost:5432/memorykana?sslmode=disable"

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
	db := setupTestDB(t)

	var exists bool
	db.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.columns
		WHERE table_name = 'migration' AND column_name = 'version'
	)`).Scan(&exists)

	if !exists {
		t.Error("expected table migration with version column to exist")
	}
}

// TODO: test column types
func Test_CreatePlayerTimesTable(t *testing.T) {
	db := setupTestDB(t)

	rows, err := db.Query(`SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'player_times'
		ORDER BY column_name ASC`)
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	i := 0
	expectedColumns := []string{"id", "player", "time"}

	for rows.Next() {
		var c string
		err := rows.Scan(&c)
		if err != nil {
			t.Fatalf("failed to read row: %v", err)
		}

		if c != expectedColumns[i] {
			t.Errorf("expected column name %s, got %v", expectedColumns[i], c)
		}
		i++
	}
}

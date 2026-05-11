package storage

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"postgres://postgres:%s@localhost:5432/memorykana?sslmode=disable",
		os.Getenv("PG_SUPERUSER_PASSWORD"))

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
		AND table_schema = current_schema()
		ORDER BY column_name ASC`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var c string
		err := rows.Scan(&c)
		if err != nil {
			t.Fatalf("failed to read row: %v", err)
		}
		got = append(got, c)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("row iteration error: %v", err)
	}

	expected := []string{"id", "player", "time"}
	if len(got) != len(expected) {
		t.Errorf("expected %d columns, got %d: %v", len(expected), len(got), got)
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("column %d: expected %q, got %q", i, expected[i], got[i])
		}
	}
}

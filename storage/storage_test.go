package storage

import (
	"errors"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	retries := 3
	baseBackoff := 1 * time.Millisecond

	t.Run("no pgvars defined", func(t *testing.T) {
		_, err := Connect(retries, baseBackoff)
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("pgvars defined", func(t *testing.T) {
		t.Setenv("PGDATABASE", "test")
		t.Setenv("PGPASSWORD", "password")
		t.Setenv("PGSSLMODE", "disable")
		t.Setenv("PGUSER", "user")

		db, err := Connect(retries, baseBackoff)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()
	})
}

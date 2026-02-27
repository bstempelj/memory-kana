package storage

import (
	"errors"
	"testing"
)

func TestConnect(t *testing.T) {
	t.Run("no pgvars defined", func(t *testing.T) {
		_, err := Connect()
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("pgvars defined", func(t *testing.T) {
		t.Setenv("PGDATABASE", "test")
		t.Setenv("PGPASSWORD", "password")
		t.Setenv("PGSSLMODE", "disable")
		t.Setenv("PGUSER", "user")

		db, err := Connect()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer db.Close()
	})
}

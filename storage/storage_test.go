package storage

import (
	"errors"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	retries := 3
	baseBackoff := 1 * time.Millisecond

	t.Run("no PG env vars defined", func(t *testing.T) {
		_, err := Connect(retries, baseBackoff)
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("missing PGUSER, PGPASSWORD and PGSSLMODE env vars", func(t *testing.T) {
		t.Setenv("PGDATABASE", "test")

		_, err := Connect(retries, baseBackoff)
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("missing PGPASSWORD AND PGSSLMODE env vars", func(t *testing.T) {
		t.Setenv("PGDATABASE", "test")
		t.Setenv("PGUSER", "user")

		_, err := Connect(retries, baseBackoff)
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("missing PGSSLMODE env var", func(t *testing.T) {
		t.Setenv("PGDATABASE", "test")
		t.Setenv("PGUSER", "user")
		t.Setenv("PGPASSWORD", "password")

		_, err := Connect(retries, baseBackoff)
		if !errors.Is(err, ErrPostgresTimeout) {
			t.Fatalf("expected %v, got %v", ErrPostgresTimeout, err)
		}
	})

	t.Run("all PG env vars defined", func(t *testing.T) {
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

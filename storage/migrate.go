package storage

import (
	"database/sql"
	"log/slog"
)

type Migration func(tx *sql.Tx) error

func createMigrationTable(tx *sql.Tx) error {
	slog.Info("creating migration table")
	_, err := tx.Exec("CREATE TABLE IF NOT EXISTS migration (version int PRIMARY KEY)")
	if err != nil {
		return err
	}

	slog.Info("inserting version 0 into migration table")
	_, err = tx.Exec("INSERT INTO migration (version) VALUES (0) ON CONFLICT DO NOTHING")
	if err != nil {
		return err
	}
	return nil
}

func createPlayerTimesTable(tx *sql.Tx) error {
	slog.Info("creating player_times table")
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS player_times(
		id SERIAL PRIMARY KEY,
		player VARCHAR(50) NOT NULL,
		"time" TIME NOT NULL
	)`)
	if err != nil {
		return err
	}
	return nil
}

// TODO: use singular for all table names

func Migrate(db *sql.DB) error {
	slog.Info("starting database migration")
	migrations := []Migration{
		createMigrationTable,
		createPlayerTimesTable,
	}
	latestVersion := len(migrations)

	var version int
	err := db.QueryRow("SELECT version FROM migration").Scan(&version)
	if err != nil {
		version = 0
	}

	slog.Info("current migration version", "version", version)
	slog.Info("latest migration version", "version", latestVersion)

	for version < latestVersion {
		slog.Info("starting transaction")
		// start transaction
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		// run migration
		err = migrations[version](tx)
		if err != nil {
			slog.Info("rolling back transaction")
			tx.Rollback()
			return err
		}

		version++

		slog.Info("updating migration version", "version", version)
		_, err = tx.Exec("UPDATE migration SET version = $1", version)
		if err != nil {
			slog.Info("rolling back transaction")
			tx.Rollback()
			return err
		}
		tx.Commit()
	}

	slog.Info("database migration complete")
	return nil
}

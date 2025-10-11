package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
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

func convertPlayerTimesTableToPlayerDuration(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE player_times
		ADD COLUMN IF NOT EXISTS duration BIGINT
	`)
	if err != nil {
		return err
	}

	rows, err := tx.Query(`
		SELECT id, "time"
		FROM player_times
		WHERE duration IS NULL
	`)
	if err != nil {
		return err
	}

	type playerTime struct {
		id int
		time time.Time
		duration time.Duration
	}

	var ptList []playerTime

	for rows.Next() {
		var pt playerTime

		err := rows.Scan(&pt.id, &pt.time)
		if err != nil {
			rows.Close()
			return err
		}

		ptList = append(ptList, pt)
	}

	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}

	utcYear := time.Unix(0, 0).UTC().Year()
	for _, pt := range ptList {
		// convert year to utc year
		pt.time = pt.time.AddDate(utcYear - pt.time.Year(), 0, 0)
		pt.duration = pt.time.Sub(time.Unix(0, 0).UTC())

		_, err = tx.Exec(`
			UPDATE player_times
			SET duration = $1
			WHERE id = $2 AND "time" = $3`,
			pt.duration, pt.id, pt.time)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		ALTER TABLE player_times
		ALTER COLUMN duration SET NOT NULL
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE player_times
		DROP COLUMN "time"
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE player_times
		RENAME TO player_duration
	`)
	if err != nil {
		return err
	}
	return nil
}

func Migrate(db *sql.DB) error {
	slog.Info("starting database migration")
	migrations := []Migration{
		createMigrationTable,
		createPlayerTimesTable,
		convertPlayerTimesTableToPlayerDuration,
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
		defer tx.Rollback()

		// run migration
		err = migrations[version](tx)
		if err != nil {
			slog.Info("rolling back transaction")
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("migration failed: %w; rollback failed: %w", err, rbErr)
			}
			return err
		}

		version++

		slog.Info("updating migration version", "version", version)
		_, err = tx.Exec("UPDATE migration SET version = $1", version)
		if err != nil {
			slog.Info("rolling back transaction")
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("migration failed: %w; rollback failed: %w", err, rbErr)
			}
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	slog.Info("database migration complete")
	return nil
}

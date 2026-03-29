package storage

import "database/sql"

type Migration func(tx *sql.Tx) error

// TODO: use singular for all table names
func createMigrationTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS migration (version int PRIMARY KEY)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO migration (version) VALUES (0)`)
	if err != nil {
		return err
	}
	return nil
}

func createPlayerTimesTable(tx *sql.Tx) error {
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

func Migrate(db *sql.DB) error {
	migrations := []Migration{
		createMigrationTable,
		createPlayerTimesTable,
	}
	latestVersion := len(migrations)

	// NOTE: if not exists version is still probably 0 (have to test this)
	var version int
	db.QueryRow("SELECT version FROM migration").Scan(&version)

	for version < latestVersion {
		// start transaction
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// run migration
		err = migrations[version](tx)
		if err != nil {
			return err
		}

		version++

		// update migration version in db
		_, err = tx.Exec("UPDATE migration SET version = $1", version)
		if err != nil {
			return err
		}
		tx.Commit()
	}
	return nil
}

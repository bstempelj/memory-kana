package storage

import (
	"database/sql"
	"math/rand/v2"
	"time"

	"github.com/bstempelj/memory-kana/hash"
)

type Migration func(tx *sql.Tx) error

// TODO: move migration code to storage/migration.go
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

// TODO: should be applied after all migrations
// and should seed the new table player_duration
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

func seedPlayerTimesTable(tx *sql.Tx) error {
	// TODO: convert to a single insert statement (faster)
	numRows := 10
	for i := 0; i < numRows; i++ {
		// between 10s and 90s
		duration := time.Duration(rand.IntN(80)+10) * time.Second
		_, err := tx.Exec(
			"insert into player_times(player, time) values ($1, $2)",
			"generated-"+hash.Random(8),
			time.Unix(0, 0).UTC().Add(duration),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func changePlayerTimesTableToPlayerDuration(tx *sql.Tx) error {
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
	defer rows.Close()

	for rows.Next() {
		var (
			ptID   int
			ptTime time.Time
		)
		err := rows.Scan(&ptID, &ptTime)
		if err != nil {
			return err
		}

		// convert year to utc year
		year := ptTime.Year()
		utcYear := time.Unix(0, 0).UTC().Year()
		ptTime = ptTime.AddDate(utcYear-year, 0, 0)

		ptDuration := ptTime.Sub(time.Unix(0, 0).UTC())

		_, err = tx.Exec(`
			UPDATE player_times
			SET duration = $1
			WHERE id = $2 AND "time" = $3`,
			ptDuration, ptID, ptTime,
		)
		if err != nil {
			return err
		}
	}

	if err = rows.Err(); err != nil {
		return err
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

	// TODO: enable after methods in storage are updated with new table name
	// _, err = tx.Exec(`
	// 	ALTER TABLE player_times
	// 	RENAME TO player_duration
	// `)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func Migrate(db *sql.DB) error {
	migrations := []Migration{
		createMigrationTable,
		createPlayerTimesTable,
		// NOTE: this is just for testing, move to a *_test.go
		//seedPlayerTimesTable,
		changePlayerTimesTableToPlayerDuration,
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

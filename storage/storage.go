package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
	_ "github.com/lib/pq"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type PlayerTime struct {
	Player string
	Time   time.Time
}

var ErrPostgresTimeout = errors.New("connection to postgres timed out")

func Connect() (*sql.DB, error) {
	var db *sql.DB
	var err error

	retries := 5
	delay := 1 * time.Second

	slog.Info("starting connection to postgres")

	for i := 0; i < retries; i++ {
		// connection string info is read from pg env vars
		db, err = sql.Open("postgres", "")
		if err != nil {
			return nil, fmt.Errorf("failed to open connection to postgres: %w", err)
		}

		if err = db.Ping(); err == nil {
			slog.Info("connection to postgres successfull")
			return db, nil
		}

		slog.Info(fmt.Sprintf("retrying after %ds...", delay / time.Second))
		time.Sleep(delay)
		delay *= 2
	}
	return nil, ErrPostgresTimeout
}

func InsertPlayerTime(db *sql.DB, playerTime time.Time) (string, error) {
	randHash := func(length int) string {
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	playerName := "guest-" + randHash(8)

	_, err := db.Exec(
		"insert into player_times(player, time) values ($1, $2)",
		playerName,
		playerTime)
	if err != nil {
		return "", err
	}
	return playerName, nil
}

func SelectPlayerTimes(db *sql.DB) ([]PlayerTime, error) {
	var playerTimes []PlayerTime

	res, err := db.Query(`select player, "time" from player_times order by "time" limit 10`)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		var playerTime PlayerTime

		err := res.Scan(&playerTime.Player, &playerTime.Time)
		if err != nil {
			return nil, err
		}

		playerTimes = append(playerTimes, playerTime)
	}

	if err = res.Err(); err != nil {
		return nil, err
	}
	return playerTimes, nil
}

func SelectPlayerTimeRank(db *sql.DB, player string) (time.Time, uint, error) {
	var playerTime time.Time
	var playerRank uint

	row := db.QueryRow(`select "time" from player_times where player = $1`, player)
	err := row.Scan(&playerTime)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, 0, errors.New("no rows returned when querying for player time")
	}

	row = db.QueryRow(`select count(1) from player_times where "time" <= $1`, playerTime)
	err = row.Scan(&playerRank)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, 0, errors.New("no rows returned when querying for player rank")
	}

	return playerTime, playerRank, nil
}

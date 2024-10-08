package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"strings"
	"time"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type PlayerTime struct {
	Player string
	Time   time.Time
}

func Connect() (*sql.DB, error) {
	// NOTE: env vars have carriage return at the end (\r) so we
	// have to trim them to make concatenation and sprintf work
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		host = "localhost"
	}
	host = strings.TrimSpace(host)

	port, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		port = "5432"
	}
	port = strings.TrimSpace(port)

	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		return nil, errors.New("missing POSTGRES_USER env var")
	}
	user = strings.TrimSpace(user)

	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		return nil, errors.New("missing POSTGRES_PASSWORD env var")
	}
	password = strings.TrimSpace(password)

	dbname, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		return nil, errors.New("missing POSTGRES_DB env var")
	}
	dbname = strings.TrimSpace(dbname)

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
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

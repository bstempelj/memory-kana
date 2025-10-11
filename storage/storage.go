package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type PlayerDuration struct {
	Player   string
	Duration time.Duration
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Connect(cfg Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)

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

func InsertPlayerDuration(db *sql.DB, duration time.Duration) (string, error) {
	randHash := func(length int) string {
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	player := "guest-" + randHash(8)

	_, err := db.Exec(
		`insert into player_times(player, "time", duration) values ($1, $2, $3)`,
		player, time.Time{}, duration.Nanoseconds())
	if err != nil {
		return "", err
	}
	return player, nil
}

func SelectPlayerDurationList(db *sql.DB) ([]PlayerDuration, error) {
	var playerDurationList []PlayerDuration

	res, err := db.Query(`
		select player, duration
		from player_times
		order by duration
		limit 10
	`)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		var playerDuration PlayerDuration

		err := res.Scan(&playerDuration.Player, &playerDuration.Duration)
		if err != nil {
			return nil, err
		}

		playerDurationList = append(playerDurationList, playerDuration)
	}

	if err = res.Err(); err != nil {
		return nil, err
	}
	return playerDurationList, nil
}

func SelectPlayerDurationAndRank(db *sql.DB, player string) (time.Duration, uint, error) {
	var (
		duration time.Duration
		rank     uint
	)

	row := db.QueryRow(`
		select duration
		from player_times
		where player = $1
	`, player)
	err := row.Scan(&duration)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return time.Duration(0), 0, errors.New("no rows returned when querying for player duration")
	}

	row = db.QueryRow(`select count(1) from player_times where duration <= $1`, duration)
	err = row.Scan(&rank)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return time.Duration(0), 0, errors.New("no rows returned when querying for player rank")
	}

	return duration, rank, nil
}

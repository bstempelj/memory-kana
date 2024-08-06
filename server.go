package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

var db *sql.DB

//go:embed assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

type PlayerTime struct {
	Player string `json:"player"`
	Time   string `json:"time"`
}

type Page struct {
	Home       bool
	Scripts    bool
	Scoreboard []PlayerTime
}

func getMenu(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFS(
		templatesFS,
		"templates/base.html",
		"templates/menu.html",
	))

	page := Page{Home: true}

	if err := t.Execute(w, page); err != nil {
		log.Fatal(err)
	}
}

func getGame(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFS(
		templatesFS,
		"templates/base.html",
		"templates/game.html",
	))

	page := Page{Scripts: true}

	if err := t.Execute(w, page); err != nil {
		log.Fatal(err)
	}
}

func getScoreboard(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFS(
		templatesFS,
		"templates/base.html",
		"templates/scoreboard.html",
	))

	scoreboard, err := dbSelectPlayerTimes(db)
	if err != nil {
		// TODO: redirect to error page
		log.Fatal(err)
	}

	page := Page{Scoreboard: scoreboard}

	if err := t.Execute(w, page); err != nil {
		log.Fatal(err)
	}
}

func postScoreboard(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	playerTime := r.FormValue("player-time")

	fmt.Println("playerTime=" + playerTime)

	err := dbInsertPlayerTime(db, playerTime)
	if err != nil {
		// TODO: redirect to error page
		log.Fatal(err)
	}

	http.Redirect(w, r, "/scoreboard", http.StatusSeeOther)
}

func connect() (*sql.DB, error) {
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

func dbInsertPlayerTime(db *sql.DB, playerTime string) error {
	randHash := func(length int) string {
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	_, err := db.Exec(
		"insert into player_times(player, time) values ($1, $2)",
		"guest-"+randHash(8),
		playerTime)
	if err != nil {
		return err
	}
	return nil
}

func dbSelectPlayerTimes(db *sql.DB) ([]PlayerTime, error) {
	var playerTimes []PlayerTime

	res, err := db.Query("select player, \"time\" from player_times")
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

func main() {
	var err error

	db, err = connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getMenu)
	mux.HandleFunc("GET /game", getGame)
	mux.HandleFunc("GET /scoreboard", getScoreboard)
	mux.HandleFunc("POST /scoreboard", postScoreboard)
	mux.Handle("GET /assets/", http.FileServer(http.FS(assetsFS)))

	port := 1234

	log.Printf("Listening on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), mux))
}

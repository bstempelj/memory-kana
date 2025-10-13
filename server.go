package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/csrf"

	"github.com/bstempelj/memory-kana/handlers"
	"github.com/bstempelj/memory-kana/storage"
)

//go:embed assets
var assets embed.FS

//go:embed templates
var templates embed.FS

var storageCfg storage.Config

func init() {
	// NOTE: env vars have carriage return at the end (\r) so we
	// have to trim them to make concatenation and sprintf work
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		host = "localhost"
	}
	storageCfg.Host = strings.TrimSpace(host)

	port, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		port = "5432"
	}
	storageCfg.Port = strings.TrimSpace(port)

	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		log.Fatal("missing POSTGRES_USER env var")
	}
	storageCfg.User = strings.TrimSpace(user)

	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		log.Fatal("missing POSTGRES_PASSWORD env var")
	}
	storageCfg.Password = strings.TrimSpace(password)

	dbname, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		log.Fatal("missing POSTGRES_DB env var")
	}
	storageCfg.Database = strings.TrimSpace(dbname)
}

func main() {
	db, err := storage.Connect(storageCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// NOTE: add new column "duration" to the "player_times" table
	{
		_, err := db.Exec(`
			ALTER TABLE player_times
			ADD COLUMN IF NOT EXISTS duration BIGINT NOT NULL`)
		if err != nil {
			log.Fatal(err)
		}
	}

	// NOTE: migrate data from time column to duration
	{
		res, err := db.Query(`
			select id, "time"
			from player_times
			where duration = 0
		`)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Close()

		for res.Next() {
			var ptID int
			var ptTime time.Time
			err := res.Scan(&ptID, &ptTime)
			if err != nil {
				log.Fatal(err)
			}

			ptDuration := ptTime.Sub(time.Unix(0, 0))

			_, err = db.Exec(
				"UPDATE player_times SET duration = $1 WHERE id = $2 AND \"time\" = $3",
				ptDuration, ptID, ptTime,
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err = res.Err(); err != nil {
			log.Fatal(err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.GetMenu(templates, db))
	mux.HandleFunc("GET /game", handlers.GetGame(templates, db))
	mux.Handle("GET /game/ws", handlers.NewWebSocketHandler(db))
	mux.HandleFunc("GET /scoreboard", handlers.GetScoreboard(templates, db))
	mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))

	port := 1234

	csrfAuthKey, ok := os.LookupEnv("CSRF_AUTH_KEY")
	if !ok {
		log.Fatal("missing CSRF_AUTH_KEY env var")
	}
	csrfAuthKey = strings.TrimSpace(csrfAuthKey)

	hostEnv, ok := os.LookupEnv("HOST_ENV")
	if !ok {
		log.Fatal("missing HOST_ENV env var")
	}
	hostEnv = strings.TrimSpace(hostEnv)

	csrfSecure := false
	if hostEnv == "prod" || hostEnv == "production" {
		csrfSecure = true
	}

	CSRF := csrf.Protect([]byte(csrfAuthKey), csrf.Secure(csrfSecure))

	log.Printf("Listening on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), CSRF(mux)))
}

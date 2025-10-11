package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"flag"
	"database/sql"

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

// TODO: move migration code to storage/migration.go
// TODO: use singular for all table names
func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		ALTER TABLE player_times
		ADD COLUMN IF NOT EXISTS duration BIGINT
	`)
	if err != nil {
		return err
	}

	res, err := db.Query(`
		SELECT id, "time"
		FROM player_times
		WHERE duration IS NULL
	`)
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		var (
			ptID int
			ptTime time.Time
		)
		err := res.Scan(&ptID, &ptTime)
		if err != nil {
			return err
		}

		// convert year to utc year
		year := ptTime.Year()
		utcYear := time.Unix(0, 0).UTC().Year()
		ptTime = ptTime.AddDate(utcYear - year, 0, 0)

		ptDuration := ptTime.Sub(time.Unix(0, 0).UTC())

		_, err = db.Exec(`
			UPDATE player_times
			SET duration = $1
			WHERE id = $2 AND "time" = $3`,
			ptDuration, ptID, ptTime,
		)
		if err != nil {
			return err
		}
	}

	if err = res.Err(); err != nil {
		return err
	}

	_, err = db.Exec(`
		ALTER TABLE player_times
		ALTER COLUMN duration SET NOT NULL
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		ALTER TABLE player_times
		DROP COLUMN "time"
	`)
	if err != nil {
		return err
	}

	// TODO: rename table player_times to player_duration
	return nil
}

func main() {
	migrate := flag.Bool("migrate", false, "run migrations")
	flag.Parse()

	db, err := storage.Connect(storageCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *migrate {
		fmt.Print("Migrating database...")
		if err := runMigrations(db); err != nil {
			fmt.Println("NOK")
			log.Fatal(err)
		}
		fmt.Println("OK")
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.GetMenu(templates, db))
	mux.HandleFunc("GET /game", handlers.GetGame(templates, db))
	mux.HandleFunc("GET /scoreboard", handlers.GetScoreboard(templates, db))
	mux.HandleFunc("POST /scoreboard", handlers.PostScoreboard(db))
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

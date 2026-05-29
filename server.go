package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"

	"github.com/bstempelj/memory-kana/handlers"
	"github.com/bstempelj/memory-kana/storage"
)

//go:embed assets
var assets embed.FS

//go:embed templates
var templates embed.FS

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "run only migrations without starting the server")
	flag.Parse()

	db, err := storage.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = storage.Migrate(db); err != nil {
		log.Fatal(err)
	}
	if *migrateOnly {
		os.Exit(0)
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

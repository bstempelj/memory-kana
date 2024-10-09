package main

import (
	"embed"
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

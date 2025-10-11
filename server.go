package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
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
		slog.Error("database connection", "err", err)
		os.Exit(1)
	}
	defer storage.CloseDB(db)

	if err = storage.Migrate(db); err != nil {
		slog.Error("database migration", "err", err)
		os.Exit(1)
	}
	if *migrateOnly {
		slog.Info("database migration finished")
		os.Exit(0)
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
		slog.Error("missing env var", "env", "CSRF_AUTH_KEY")
		os.Exit(1)
	}
	csrfAuthKey = strings.TrimSpace(csrfAuthKey)

	hostEnv, ok := os.LookupEnv("HOST_ENV")
	if !ok {
		slog.Error("missing env var", "env", "HOST_ENV")
		os.Exit(1)
	}
	hostEnv = strings.TrimSpace(hostEnv)

	csrfSecure := hostEnv == "prod" || hostEnv == "production"
	CSRF := csrf.Protect([]byte(csrfAuthKey), csrf.Secure(csrfSecure))

	slog.Info("server starting", "port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), CSRF(mux))
	if err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

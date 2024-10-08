package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/csrf"

	"github.com/bstempelj/memory-kana/storage"
)

//go:embed assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

type Page struct {
	Home       bool
	Scripts    bool
	// todo: define template type with time=string
	Scoreboard []storage.PlayerTime
	CSRFToken  string

	// tmp
	Name string
	Time time.Time
	Rank uint
}

func getMenu(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func getGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFS(
			templatesFS,
			"templates/base.html",
			"templates/game.html",
		))

		page := Page{
			Scripts:   true,
			CSRFToken: csrf.Token(r),
		}

		if err := t.Execute(w, page); err != nil {
			log.Fatal(err)
		}
	}
}

func getScoreboard(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		player := r.URL.Query().Get("p")

		funcMap := template.FuncMap{
			"formatTime": func(t time.Time) string {
				return fmt.Sprintf("%02d:%02d", t.Minute(), t.Second())
			},
		}

		t, err := template.New("base.html").Funcs(funcMap).ParseFS(
			templatesFS,
			"templates/base.html",
			"templates/scoreboard.html",
		)

		if err != nil {
			log.Fatal(err)
		}

		scoreboard, err := storage.SelectPlayerTimes(db)
		if err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}

		page := Page{
			Scoreboard: scoreboard,
		}

		if player != "" {
			playerTime, playerRank, err := storage.SelectPlayerTimeRank(db, player)
			if err != nil {
				// TODO: redirect to error page
				log.Fatal(err)
			}

			page.Name = player
			page.Time = playerTime
			page.Rank = playerRank
		}

		if err := t.Execute(w, page); err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}
	}
}

func postScoreboard(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		playerTime, err := time.Parse("04:05", r.FormValue("player-time"))
		if err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}

		playerName, err := storage.InsertPlayerTime(db, playerTime)
		if err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}

		http.Redirect(w, r, "/scoreboard?p="+playerName, http.StatusSeeOther)
	}
}

func main() {
	db, err := storage.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getMenu(db))
	mux.HandleFunc("GET /game", getGame(db))
	mux.HandleFunc("GET /scoreboard", getScoreboard(db))
	mux.HandleFunc("POST /scoreboard", postScoreboard(db))
	mux.Handle("GET /assets/", http.FileServer(http.FS(assetsFS)))

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

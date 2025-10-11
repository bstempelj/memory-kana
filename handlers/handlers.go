package handlers

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/bstempelj/memory-kana/storage"
	"github.com/gorilla/csrf"
	"log"
	"net/http"
	"text/template"
	"time"
)

type Page struct {
	Home    bool
	Scripts bool

	Scoreboard []storage.PlayerDuration
	CSRFToken  string
	Kana       string

	// tmp
	Name string
	Duration time.Duration
	Rank uint
}

func GetMenu(templateFS embed.FS, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFS(
			templateFS,
			"templates/base.html",
			"templates/menu.html",
		))

		page := Page{Home: true}

		if err := t.Execute(w, page); err != nil {
			log.Fatal(err)
		}
	}
}

func GetGame(templateFS embed.FS, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFS(
			templateFS,
			"templates/base.html",
			"templates/game.html",
		))

		kana := r.URL.Query().Get("kana")

		if kana == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		page := Page{
			Scripts:   true,
			Kana:      kana,
			CSRFToken: csrf.Token(r),
		}

		if err := t.Execute(w, page); err != nil {
			log.Fatal(err)
		}
	}
}

func GetScoreboard(templateFS embed.FS, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		player := r.URL.Query().Get("p")

		funcMap := template.FuncMap{
			"formatDuration": func(t time.Duration) string {
				minutes := int(t.Minutes()) % 60
				seconds := int(t.Seconds()) % 60
				return fmt.Sprintf("%02d:%02d", minutes, seconds)
			},
		}

		t, err := template.New("base.html").Funcs(funcMap).ParseFS(
			templateFS,
			"templates/base.html",
			"templates/scoreboard.html",
		)

		if err != nil {
			log.Fatal(err)
		}

		scoreboard, err := storage.SelectPlayerDurationList(db)
		if err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}

		page := Page{
			Scoreboard: scoreboard,
		}

		if player != "" {
			duration, rank, err := storage.SelectPlayerDurationAndRank(db, player)
			if err != nil {
				// TODO: redirect to error page
				log.Fatal(err)
			}

			page.Name = player
			page.Duration = duration
			page.Rank = rank
		}

		if err := t.Execute(w, page); err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}
	}
}

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
	// todo: define template type with time=string
	Scoreboard []storage.PlayerTime
	CSRFToken  string

	// tmp
	Name string
	Time time.Time
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

		page := Page{
			Scripts:   true,
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
			"formatTime": func(t time.Time) string {
				return fmt.Sprintf("%02d:%02d", t.Minute(), t.Second())
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

func PostScoreboard(db *sql.DB) http.HandlerFunc {
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

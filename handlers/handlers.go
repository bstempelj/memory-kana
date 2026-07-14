package handlers

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/bstempelj/memory-kana/storage"
	"github.com/gorilla/csrf"
	"log/slog"
	"net/http"
	"os"
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
	Name     string
	Duration time.Duration
	Rank     uint
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
			slog.Error("html template rendering", "page", "menu", "err", err)
			os.Exit(1)
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
			slog.Error("html template rendering", "page", "game", "err", err)
			os.Exit(1)
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
			slog.Error("html template parsing", "page", "scoreboard", "err", err)
			os.Exit(1)
		}

		scoreboard, err := storage.SelectPlayerDurationList(db)
		if err != nil {
			// TODO: redirect to error page
			slog.Error("scoreboard retrieval", "page", "scoreboard", "err", err)
			os.Exit(1)
		}

		page := Page{
			Scoreboard: scoreboard,
		}

		if player != "" {
			duration, rank, err := storage.SelectPlayerDurationAndRank(db, player)
			if err != nil {
				// TODO: redirect to error page
				slog.Error("player duration and rank retrieval", "page", "scoreboard", "err", err)
				os.Exit(1)
			}

			page.Name = player
			page.Duration = duration
			page.Rank = rank
		}

		if err := t.Execute(w, page); err != nil {
			// TODO: redirect to error page
			slog.Error("html template rendering", "page", "scoreboard", "err", err)
			os.Exit(1)
		}
	}
}

func PostScoreboard(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			// TODO: redirect to error page
			slog.Error("html form parsing", "page", "scoreboard", "err", err)
			os.Exit(1)
		}

		playerTime, err := time.Parse("04:05", r.FormValue("player-time"))
		if err != nil {
			// TODO: redirect to error page
			slog.Error("player time parsing", "page", "scoreboard", "err", err)
			os.Exit(1)
		}

		// remove 1 year since zero year is january 1, year 1
		duration := playerTime.Sub((time.Time{}).AddDate(-1, 0, 0))

		playerName, err := storage.InsertPlayerDuration(db, duration)
		if err != nil {
			// TODO: redirect to error page
			slog.Error("player time insertion", "page", "scoreboard", "err", err)
			os.Exit(1)
		}

		http.Redirect(w, r, "/scoreboard?p="+playerName, http.StatusSeeOther)
	}
}

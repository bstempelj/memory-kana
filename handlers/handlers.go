package handlers

import (
	"strconv"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"
	"sync"

	"github.com/bstempelj/memory-kana/storage"
	"github.com/bstempelj/memory-kana/utils"

	"github.com/gorilla/csrf"
)

type TimerResponse struct {
	TimerID string `json:"timerID"`
	StartTime int64  `json:"startTime"`
	StopTime  *int64 `json:"stopTime,omitempty"`
}

type SyncTimeMap struct {
	sync.RWMutex
	timeMap map[string]int64 // timerID: startTime
}

func (t *SyncTimeMap) Store(timerID string, startTime int64) {
	t.Lock()
	defer t.Unlock()
	t.timeMap[timerID] = startTime
}

func (t *SyncTimeMap) Load(timerID string) (int64, bool) {
	t.RLock()
	defer t.RUnlock()
	startTime, ok := t.timeMap[timerID]
	return startTime, ok
}

func (t *SyncTimeMap) Delete(timerID string) {
	t.Lock()
	defer t.Unlock()
	delete(t.timeMap, timerID)
}

var globalTimers SyncTimeMap

func init() {
	globalTimers.timeMap = make(map[string]int64)
}

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

func GetTimer(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")

	var response TimerResponse

	switch action {
	case "start":
		timerID := utils.RandHash(8)
		startTime := time.Now().UnixMilli()

		globalTimers.Store(timerID, startTime)

		response = TimerResponse{
			TimerID: timerID,
			StartTime: startTime,
		}
	case "stop":
		timerID := r.URL.Query().Get("tid")

		startTime, ok := globalTimers.Load(timerID)
		if !ok {
			log.Fatal("timer with id " + timerID + " does not exists")
		}

		stopTime := time.Now().UnixMilli()

		response = TimerResponse{
			TimerID: timerID,
			StartTime: startTime,
			StopTime:  &stopTime,
		}

		globalTimers.Delete(timerID)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func GetScoreboard(templateFS embed.FS, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		player := r.URL.Query().Get("p")

		funcMap := template.FuncMap{
			"formatTime": func(t time.Time) string {
				return fmt.Sprintf("%02d:%02d:%02d", t.Minute(), t.Second(), t.Nanosecond() / 1_000_000)
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

		unixMilli, err := strconv.ParseInt(r.FormValue("player-time"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		playerTime := time.UnixMilli(unixMilli)

		playerName, err := storage.InsertPlayerTime(db, playerTime)
		if err != nil {
			// TODO: redirect to error page
			log.Fatal(err)
		}

		http.Redirect(w, r, "/scoreboard?p="+playerName, http.StatusSeeOther)
	}
}

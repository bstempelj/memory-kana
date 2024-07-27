package main

import (
    "net/http"
    "log"
    "fmt"
    "embed"
    "text/template"
)

//go:embed assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

var scoreboard []PlayerScore

type PlayerScore struct {
    Player string `json:"player"`
    Score string `json:"score"`
}

func getMenu(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFS(
        templatesFS,
        "templates/base.html",
        "templates/menu.html",
    ))
    t.Execute(w, nil)
}

func getGame(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFS(
        templatesFS,
        "templates/base.html",
        "templates/game.html",
    ))
    t.Execute(w, nil)
}

func getScoreboard(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFS(
        templatesFS,
        "templates/base.html",
        "templates/scoreboard.html",
    ))
    t.Execute(w, scoreboard)
}

func postScoreboard(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    scoreboard = append(scoreboard, PlayerScore{
        Player: r.FormValue("player"),
        Score: r.FormValue("score"),
    })

    log.Println(scoreboard)

    http.Redirect(w, r, "/scoreboard", http.StatusSeeOther)
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /", getMenu)
    mux.HandleFunc("GET /game", getGame)
    mux.HandleFunc("GET /scoreboard", getScoreboard)
    mux.HandleFunc("POST /scoreboard", postScoreboard)
    mux.Handle("GET /assets/", http.FileServer(http.FS(assetsFS)))

    port := 1234

    log.Printf("Listening on port %v", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), mux))
}

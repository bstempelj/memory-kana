package main

import (
    "net/http"
    "log"
    "os"
    "text/template"
)

type PlayerScore struct {
    Player string `json:"player"`
    Score string `json:"score"`
}

var scoreboard []PlayerScore

func getScoreboard(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("../scoreboard.html"))
    tmpl.Execute(w, scoreboard)
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

func newFileServer(clientPath string) http.Handler {
    _, err := os.Stat(clientPath)
    if err != nil {
        log.Fatal(err)
    }

    return http.FileServer(http.Dir(clientPath))
}

func main() {
    mux := http.NewServeMux()
    mux.Handle("/", newFileServer(".."))
    mux.HandleFunc("GET /scoreboard", getScoreboard)
    mux.HandleFunc("POST /scoreboard", postScoreboard)

    log.Printf("Listening on port 1234")
    err := http.ListenAndServe(":1234", mux)
    if err != nil {
        log.Fatal(err)
    }
}

package main

import (
    "net/http"
    "log"
    "os"
    "sort"
    "text/template"
    "encoding/json"
)

type PlayerScore struct {
    Player string `json:"player"`
    Score int `json:"score"`
}

var scoreboard []PlayerScore

func getScoreboard(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("../scoreboard.html"))
    tmpl.Execute(w, scoreboard)
}

func postScoreboard(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)

    var playerScore PlayerScore
    err := decoder.Decode(&playerScore)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    scoreboard = append(scoreboard, playerScore)

    w.WriteHeader(http.StatusOK) 
}

func init() {
    scoreboard = []PlayerScore{
        { "Norman Allison", 5 },
        { "Alyssa Cohen", 99 },
        { "Ann Elliott", 84 },
        { "Conrad Powell", 88 },
        { "Spencer Paul", 69 },
        { "Cedric Roy", 64 },
        { "Angel Sharp", 68 },
        { "Louis Evans", 45 },
        { "Terri Terry", 11 },
        { "Trevor River", 9 },
    }

    sort.Slice(scoreboard, func(i, j int) bool {
        return scoreboard[i].Score > scoreboard[j].Score
    })
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
    mux.HandleFunc("/scoreboard", getScoreboard)
    mux.HandleFunc("POST /api/scoreboard", postScoreboard)

    log.Printf("Listening on port 1234")
    err := http.ListenAndServe(":1234", mux)
    if err != nil {
        log.Fatal(err)
    }
}

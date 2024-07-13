package main

import (
    "net/http"
    "log"
    "os"
    "encoding/json"
    "sort"
)

type PlayerScore struct {
    Player string `json:"player"`
    Score int `json:"score"`
}

var scoreboard []PlayerScore

func getScoreboard(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    err := json.NewEncoder(w).Encode(scoreboard)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
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
    mux.HandleFunc("GET /api/scoreboard", getScoreboard)

    log.Printf("Listening on port 1234")
    err := http.ListenAndServe(":1234", mux)
    if err != nil {
        log.Fatal(err)
    }
}

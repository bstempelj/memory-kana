package main

import (
    "net/http"
    "log"
    "os"
    "fmt"
    "text/template"
)

const templatesPath = "../templates/"

type PlayerScore struct {
    Player string `json:"player"`
    Score string `json:"score"`
}

var scoreboard []PlayerScore

func getMenu(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFiles(
        templatesPath + "base.html",
        templatesPath + "menu.html",
    ))
    t.Execute(w, nil)
}

func getGame(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFiles(
        templatesPath + "base.html",
        templatesPath + "game.html",
    ))
    t.Execute(w, nil)
}

func getScoreboard(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFiles(
        templatesPath + "base.html",
        templatesPath + "scoreboard.html",
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
    // test that assets are available
    assetsPath := "../assets"
    _, err := os.Stat(assetsPath)
    if err != nil {
        log.Fatal(err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("GET /", getMenu)
    mux.HandleFunc("GET /game", getGame)
    mux.HandleFunc("GET /scoreboard", getScoreboard)
    mux.HandleFunc("POST /scoreboard", postScoreboard)
    mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath))))

    port := 1234

    log.Printf("Listening on port %v", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), mux))
}

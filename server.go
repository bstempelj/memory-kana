package main

import (
    "net/http"
    "log"
    "os"
    "fmt"
    "embed"
    "text/template"
    "database/sql"
    _ "github.com/lib/pq"
    "time"
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

func connect() (*sql.DB, error) {
    var (
        host = "postgres"
        //port = os.Getenv("POSTGRES_PORT")
        port = 5432
        user = os.Getenv("POSTGRES_USER")
        password = os.Getenv("POSTGRES_PASSWORD")
        dbname = os.Getenv("POSTGRES_DB")
    )

    psqlInfo := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }
    return db, nil
}

func insertPlayerTime(db *sql.DB) error {
    _, err := db.Exec("insert into player_time(player, time) values ($1, $2)", "player1", time.Now())
    if err != nil {
        return err
    }
    return nil
}

func main() {
    db, err := connect()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    err = insertPlayerTime(db)
    if err != nil {
        log.Fatal(err)
    }

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

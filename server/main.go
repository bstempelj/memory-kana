package main

import (
    "net/http"
    "log"
    "os"
)

func main() {
    clientPath := ".."

    _, err := os.Stat(clientPath)
    if err != nil {
        log.Fatal(err)
    }

    fs := http.FileServer(http.Dir(clientPath))

    mux := http.NewServeMux()
    mux.Handle("/", fs)

    err = http.ListenAndServe(":1234", mux)
    if err != nil {
        log.Fatal(err)
    }
}

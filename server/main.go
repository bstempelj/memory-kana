package main

import (
    "net/http"
    "log"
    "os"
)


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

    err := http.ListenAndServe(":1234", mux)
    if err != nil {
        log.Fatal(err)
    }
}

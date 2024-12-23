package main

import (
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	mux := http.NewServeMux()
	l := storage.NewLinksStorage()
	mux.HandleFunc("POST /", postlink.New(l).Handle)
	mux.HandleFunc("GET /", getlink.New(l).Handle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/handlers/getLink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postLink"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	mux := http.NewServeMux()
	l := storage.NewLinksStorage()
	mux.HandleFunc("POST /", postLink.New(l).Handle)
	mux.HandleFunc("GET /", getLink.New(l).Handle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/handlers/get_link"
	"github.com/ruslantos/go-shortener-service/internal/handlers/post_link"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	mux := http.NewServeMux()
	l := storage.NewLinksStorage()
	mux.HandleFunc("POST /", post_link.New(l).Handle)
	mux.HandleFunc("GET /", get_link.New(l).Handle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

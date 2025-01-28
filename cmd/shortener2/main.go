package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/file"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink2"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink2"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shorten2"
	"github.com/ruslantos/go-shortener-service/internal/middleware"
	"github.com/ruslantos/go-shortener-service/internal/middleware/compress2"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	config.ParseFlags()

	fileProducer, err := fileClient.NewProducer(config.FileStoragePath + "links")
	fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath + "links")

	l := storage.NewLinksStorage(fileConsumer)
	err = l.InitLinkMap()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(compress.GzipMiddleware, compress.GzipMiddleware2, middleware.LoggerChi(logger))

	r.Post("/", postlink.New(l, fileProducer).Handle)
	r.Get("/{link}", getlink.New(l).Handle)
	r.Post("/api/shorten", shorten.New(l, fileProducer).Handle)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		panic(err)
	}
}

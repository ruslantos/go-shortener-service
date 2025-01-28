package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shorten"
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
	if err != nil {
		panic(err)
	}

	fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath + "links")
	if err != nil {
		panic(err)
	}

	l := storage.NewLinksStorage(fileConsumer)
	err = l.InitLinkMap()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(compress.GzipMiddleware, compress.GzipMiddleware2, middleware.LoggerChi(logger))

	postLinkHandler := postlink.New(l, fileProducer)
	getLinkHandler := getlink.New(l)
	shortenHandler := shorten.New(l, fileProducer)

	r.Post("/", postLinkHandler.Handle)
	r.Get("/{link}", getLinkHandler.Handle)
	r.Post("/api/shorten", shortenHandler.Handle)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		panic(err)
	}
}

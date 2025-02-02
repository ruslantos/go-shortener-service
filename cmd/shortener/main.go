package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/handlers/getlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/ping"
	"github.com/ruslantos/go-shortener-service/internal/handlers/postlink"
	"github.com/ruslantos/go-shortener-service/internal/handlers/shorten"
	"github.com/ruslantos/go-shortener-service/internal/links"
	"github.com/ruslantos/go-shortener-service/internal/middleware/compress"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/storage"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	config.ParseFlags()

	//var db *sqlx.DB
	//if config.IsDatabaseExist {
	db, err := sqlx.Open("pgx", config.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//}

	fileProducer, err := fileClient.NewProducer(config.FileStoragePath)
	if err != nil {
		panic(err)
	}

	fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath)
	if err != nil {
		panic(err)
	}

	linkRepo := storage.NewLinksStorage(fileConsumer, db)

	if config.IsDatabaseExist {
		err = linkRepo.InitDB()
		if err != nil {
			panic(err)
		}
	} else {
		err = linkRepo.InitLinkMap()
		if err != nil {
			panic(err)
		}
	}

	linkService := links.NewLinkService(linkRepo, fileProducer)

	postLinkHandler := postlink.New(linkService)
	getLinkHandler := getlink.New(linkService)
	shortenHandler := shorten.New(linkService)
	pingHandler := ping.New(linkService)

	r := chi.NewRouter()
	r.Use(compress.GzipMiddlewareWriter, compress.GzipMiddlewareReader, logger.LoggerChi(log))
	r.Post("/", postLinkHandler.Handle)
	r.Get("/{link}", getLinkHandler.Handle)
	r.Post("/api/shorten", shortenHandler.Handle)
	r.Get("/ping", pingHandler.Handle)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		panic(err)
	}
}

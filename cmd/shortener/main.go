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
	"github.com/ruslantos/go-shortener-service/internal/handlers/shortenbatch"
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

	var db *sqlx.DB
	if config.IsDatabaseExist {
		db, err = sqlx.Open("pgx", config.DatabaseDsn)
		if err != nil {
			logger.GetLogger().Fatal("cannot connect to database", zap.Error(err))
		}
		defer db.Close()
	}

	var fileProducer *fileClient.Producer
	var fileConsumer *fileClient.Consumer
	if config.IsFileExist {
		fileProducer, err = fileClient.NewProducer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file producer", zap.Error(err))
		}

		fileConsumer, err = fileClient.NewConsumer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file consumer", zap.Error(err))
		}
	}

	linkDBRepo := storage.NewLinksStorage(fileConsumer, db)
	linkMapRepo := storage.NewMapLinksStorage(fileConsumer)

	if config.IsDatabaseExist {
		err = linkDBRepo.InitDB()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
	} else {
		err = linkMapRepo.InitLinkMap()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize link map", zap.Error(err))
		}
	}

	linkService := links.NewLinkService(linkDBRepo, linkMapRepo, fileProducer)

	postLinkHandler := postlink.New(linkService)
	getLinkHandler := getlink.New(linkService)
	shortenHandler := shorten.New(linkService)
	pingHandler := ping.New(linkService)
	shortenBatchHandler := shortenbatch.New(linkService)

	r := chi.NewRouter()
	r.Use(compress.GzipMiddlewareWriter, compress.GzipMiddlewareReader, logger.LoggerChi(log))
	r.Post("/", postLinkHandler.Handle)
	r.Get("/{link}", getLinkHandler.Handle)
	r.Post("/api/shorten", shortenHandler.Handle)
	r.Get("/ping", pingHandler.Handle)
	r.Post("/api/shorten/batch", shortenBatchHandler.Handle)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		logger.GetLogger().Fatal("cannot start server", zap.Error(err))
	}
}

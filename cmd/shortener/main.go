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
	"github.com/ruslantos/go-shortener-service/internal/handlers/userurls"
	"github.com/ruslantos/go-shortener-service/internal/links"
	"github.com/ruslantos/go-shortener-service/internal/middleware/compress"
	"github.com/ruslantos/go-shortener-service/internal/middleware/cookie"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/storage"
	"github.com/ruslantos/go-shortener-service/internal/storage/mapfile"
	"github.com/ruslantos/go-shortener-service/internal/user"
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

	userService := user.NewUserService()

	var linkService links.LinkService

	if config.IsDatabaseExist {
		linksRepo := storage.NewLinksStorage(db)
		err = linksRepo.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
		linkService = *links.NewLinkService(linksRepo, userService)
	} else {
		linksRepo := mapfile.NewMapLinksStorage(fileConsumer, fileProducer)
		err = linksRepo.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize link map", zap.Error(err))
		}
		linkService = *links.NewLinkService(linksRepo, userService)

	}

	postLinkHandler := postlink.New(&linkService)
	getLinkHandler := getlink.New(&linkService)
	shortenHandler := shorten.New(&linkService)
	pingHandler := ping.New(&linkService)
	shortenBatchHandler := shortenbatch.New(&linkService)
	userurlsHandler := userurls.New(&linkService)

	r := chi.NewRouter()

	r.Use(compress.GzipMiddlewareWriter, compress.GzipMiddlewareReader, logger.LoggerChi(log), cookie.AuthMiddleware)
	r.Post("/", postLinkHandler.Handle)
	r.Get("/{link}", getLinkHandler.Handle)
	r.Post("/api/shorten", shortenHandler.Handle)
	r.Get("/ping", pingHandler.Handle)
	r.Post("/api/shorten/batch", shortenBatchHandler.Handle)
	r.Get("/api/user/urls", userurlsHandler.Handle)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		logger.GetLogger().Fatal("cannot start server", zap.Error(err))
	}
}

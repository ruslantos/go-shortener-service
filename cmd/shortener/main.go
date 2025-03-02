package main

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/service"
	"github.com/ruslantos/go-shortener-service/internal/storage"
	"github.com/ruslantos/go-shortener-service/internal/storage/mapfile"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	config.ParseFlags()

	var linkService service.LinkService
	var linkStorage service.LinksStorage

	cfg := storage.Load()
	switch cfg.StorageType {
	case "file":
		fileProducer, err := fileClient.NewProducer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file producer", zap.Error(err))
		}
		fileConsumer, err := fileClient.NewConsumer(config.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file consumer", zap.Error(err))
		}

		linkStorage = mapfile.NewMapLinksStorage(fileConsumer, fileProducer)
		err = linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize link map", zap.Error(err))
		}
	case "postgres":
		db, err := sqlx.Open("pgx", config.DatabaseDsn)
		if err != nil {
			logger.GetLogger().Fatal("cannot connect to database", zap.Error(err))
		}
		defer db.Close()

		linkStorage = storage.NewLinksStorage(db)
		err = linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
	default:
		logger.GetLogger().Fatal("unknown storage type", zap.String("storageType", cfg.StorageType))
	}

	linkService = *service.NewLinkService(linkStorage)

	r := setupRouter(linkService, log)

	go linkService.StartDeleteWorker(context.Background())

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		logger.GetLogger().Fatal("cannot start server", zap.Error(err))
	}
}

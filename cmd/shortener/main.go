package main

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/links"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
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

	var linkService links.LinkService

	if config.IsDatabaseExist {
		linksRepo := storage.NewLinksStorage(db)
		err = linksRepo.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
		linkService = *links.NewLinkService(linksRepo)
		// запускаем воркер удаления ссылок
		go linkService.StartDeleteWorker(context.Background())
	} else {
		linksRepo := mapfile.NewMapLinksStorage(fileConsumer, fileProducer)
		err = linksRepo.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize link map", zap.Error(err))
		}
		linkService = *links.NewLinkService(linksRepo)

	}

	r := setupRouter(linkService, log)

	err = http.ListenAndServe(config.FlagServerPort, r)
	if err != nil {
		logger.GetLogger().Fatal("cannot start server", zap.Error(err))
	}
}

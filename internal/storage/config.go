package storage

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	flags "github.com/ruslantos/go-shortener-service/internal/config"
	fileClient "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/service"
	"github.com/ruslantos/go-shortener-service/internal/storage/filestorage"
	"github.com/ruslantos/go-shortener-service/internal/storage/mapstorage"
)

// Config содержит конфигурационные параметры для хранилища.
type Config struct {
	StorageType string
}

// Load загружает конфигурацию хранилища на основе флагов.
func Load(flags flags.Config) Config {
	var config Config

	if flags.IsDatabaseExist {
		config.StorageType = "postgres"
		return config
	}

	if flags.IsFileExist {
		config.StorageType = "file"
		return config
	}

	config.StorageType = "map"
	return config
}

func Get(cfg flags.Config) service.LinksStorage {
	storageCfg := Load(cfg)
	var linkStorage service.LinksStorage

	switch storageCfg.StorageType {
	case "map":
		linkStorage = mapstorage.NewMapStorage()
	case "file":
		fileProducer, err := fileClient.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file producer", zap.Error(err))
		}
		fileConsumer, err := fileClient.NewConsumer(cfg.FileStoragePath)
		if err != nil {
			logger.GetLogger().Fatal("cannot create file consumer", zap.Error(err))
		}

		linkStorage = filestorage.NewFileStorage(fileConsumer, fileProducer)
		err = linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize file storage", zap.Error(err))
		}
	case "postgres":
		db := getDB(cfg.DatabaseDsn)

		linkStorage = NewLinksStorage(db)
		err := linkStorage.InitStorage()
		if err != nil {
			logger.GetLogger().Fatal("cannot initialize database", zap.Error(err))
		}
	default:
		logger.GetLogger().Fatal("unknown storage type", zap.String("storageType", storageCfg.StorageType))
	}

	return linkStorage
}

func getDB(dsn string) *sqlx.DB {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		logger.GetLogger().Fatal("cannot connect to database", zap.Error(err))
	}
	defer db.Close()

	return db
}

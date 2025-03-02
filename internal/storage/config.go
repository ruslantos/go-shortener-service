package storage

import (
	flags "github.com/ruslantos/go-shortener-service/internal/config"
)

type Config struct {
	StorageType string
}

func Load() Config {
	var config Config

	if flags.IsDatabaseExist {
		config.StorageType = "postgres"
		return config
	}

	if flags.IsFileExist {
		config.StorageType = "file"
		return config
	}

	return config
}

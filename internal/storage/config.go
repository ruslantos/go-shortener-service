package storage

import (
	flags "github.com/ruslantos/go-shortener-service/internal/config"
)

// Config содержит конфигурационные параметры для хранилища.
type Config struct {
	StorageType string
}

// Load загружает конфигурацию хранилища на основе флагов.
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

	config.StorageType = "map"
	return config
}

package storage

import (
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib" // драйвер pgx
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flags "github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/storage/filestorage"
	"github.com/ruslantos/go-shortener-service/internal/storage/mapstorage"
)

type dbMock struct{}

func TestGet_MapStorage(t *testing.T) {
	cfg := flags.Config{
		IsDatabaseExist: false,
		IsFileExist:     false,
	}

	storage := Get(cfg)

	assert.NotNil(t, storage)
	_, ok := storage.(*mapstorage.LinksStorage)
	assert.True(t, ok)
}

func TestGet_FileStorage(t *testing.T) {
	cfg := flags.Config{
		IsDatabaseExist: false,
		IsFileExist:     true,
		FileStoragePath: "/tmp/testfile.json",
	}

	storage := Get(cfg)

	assert.NotNil(t, storage)
	_, ok := storage.(*filestorage.LinksStorage)
	assert.True(t, ok)
}

func TestLoad_StorageConfig(t *testing.T) {
	tests := []struct {
		name           string
		cfg            flags.Config
		expectedType   string
		expectFile     bool
		expectDatabase bool
	}{
		{
			name: "database exists",
			cfg: flags.Config{
				IsDatabaseExist: true,
			},
			expectedType:   "postgres",
			expectDatabase: true,
		},
		{
			name: "file exists",
			cfg: flags.Config{
				IsFileExist: true,
			},
			expectedType: "file",
			expectFile:   true,
		},
		{
			name:         "default map storage",
			cfg:          flags.Config{},
			expectedType: "map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Load(tt.cfg)
			require.Equal(t, tt.expectedType, config.StorageType)
		})
	}
}

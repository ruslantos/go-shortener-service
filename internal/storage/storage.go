package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type LinksStorage struct {
	linksMap map[string]string
	mutex    *sync.Mutex
	db       *sqlx.DB
}

func NewLinksStorage(db *sqlx.DB) *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]string),
		mutex:    &sync.Mutex{},
		db:       db,
	}
}

func (l LinksStorage) AddLink(raw models.Links) error {
	_, err := l.db.ExecContext(context.Background(),
		"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", raw.ShortURL, raw.OriginalURL)
	if err != nil {
		return err
	}
	return nil
}

func (l LinksStorage) GetLink(value string) (string, bool, error) {
	row := l.db.QueryRowContext(context.Background(),
		"SELECT original_url FROM links where short_url = $1 LIMIT 1", value)
	var long string
	err := row.Scan(&long)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return long, true, nil
}

func (l LinksStorage) Ping() error {
	if l.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	err := l.db.Ping()
	if err != nil {
		logger.GetLogger().Error(err.Error())
	}

	return err
}

func (l LinksStorage) InitDB() error {
	_, err := l.db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS links(short_url TEXT,original_url TEXT);`)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	return nil
}

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type file interface {
	ReadEvents() ([]*fileJob.Event, error)
}

type LinksStorage struct {
	linksMap map[string]string
	mutex    *sync.Mutex
	file     file
	db       *sqlx.DB
}

func NewLinksStorage(file file, db *sqlx.DB) *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]string),
		mutex:    &sync.Mutex{},
		file:     file,
		db:       db,
	}
}

func (l LinksStorage) AddLink(link models.Links) (string, error) {
	if !config.IsDatabaseExist {
		return l.getShortValue(link.OriginalURL), nil
	}

	_, err := l.db.ExecContext(context.Background(),
		"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", link.ShortURL, link.OriginalURL)
	if err != nil {
		return "", err
	}
	return link.ShortURL, nil
}

func (l LinksStorage) GetLink(value string) (string, bool, error) {
	if !config.IsDatabaseExist {
		result, ok := l.linksMap[value]
		return result, ok, nil
	}

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

func (l LinksStorage) getShortValue(raw string) string {
	l.mutex.Lock()
	count := len(l.linksMap)
	short := strconv.Itoa(count + 1)
	l.linksMap[short] = raw
	l.mutex.Unlock()

	return short
}

func (l LinksStorage) InitLinkMap() error {
	rows, err := l.file.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = row.OriginalURL
	}
	logger.GetLogger().Info("Links map initialized")
	return nil
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

	// убрать после отказа от файла и мапы
	for k, v := range l.linksMap {
		_, err := l.db.ExecContext(context.Background(),
			"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", k, v)
		if err != nil {
			logger.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}

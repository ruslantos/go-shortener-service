package storage

import (
	"fmt"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	mid "github.com/ruslantos/go-shortener-service/internal/middleware/logger"
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

func (l LinksStorage) AddLink(raw string) string {
	return l.getShortValue(raw)
}

func (l LinksStorage) GetLink(value string) (string, bool) {
	result, ok := l.linksMap[value]
	return result, ok
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
	mid.GetLogger().Info("Links map initialized")
	return nil
}

func (l LinksStorage) Ping() error {
	if l.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	err := l.db.Ping()
	if err != nil {
		mid.GetLogger().Error(err.Error())
	}

	return err
}

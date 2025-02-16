package storage

import (
	"context"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type file interface {
	ReadEvents() ([]*fileJob.Event, error)
}

type MapLinksStorage struct {
	linksMap map[string]models.Link
	mutex    *sync.Mutex
	file     file
	db       *sqlx.DB
}

func NewMapLinksStorage(file file) *MapLinksStorage {
	return &MapLinksStorage{
		linksMap: make(map[string]models.Link),
		mutex:    &sync.Mutex{},
		file:     file,
	}
}

func (l MapLinksStorage) AddLink(link models.Link) (models.Link, error) {
	l.addLinksToMap([]models.Link{link})
	return link, nil
}

func (l MapLinksStorage) AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	l.addLinksToMap(links)
	return links, nil
}

func (l MapLinksStorage) GetLink(value string) (string, bool, error) {
	result, ok := l.linksMap[value]
	return result.OriginalURL, ok, nil
}

func (l MapLinksStorage) addLinksToMap(links []models.Link) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, v := range links {
		l.linksMap[v.ShortURL] = v
	}
}

func (l MapLinksStorage) InitLinkMap() error {
	rows, err := l.file.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = models.Link{ShortURL: row.OriginalURL, OriginalURL: row.OriginalURL, CorrelationID: row.ID}
	}
	logger.GetLogger().Info("Link map initialized")
	return nil
}

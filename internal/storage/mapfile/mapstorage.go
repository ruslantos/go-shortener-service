package mapfile

import (
	"context"
	"errors"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type FileConsumer interface {
	ReadEvents() ([]*fileJob.Event, error)
}

type FileProducer interface {
	WriteEvent(event *fileJob.Event) error
}

type LinksStorage struct {
	linksMap     map[string]models.Link
	mutex        *sync.Mutex
	fileConsumer FileConsumer
	fileProducer FileProducer
	db           *sqlx.DB
}

func NewMapLinksStorage(fileConsumer FileConsumer, fileProducer FileProducer) *LinksStorage {
	return &LinksStorage{
		linksMap:     make(map[string]models.Link),
		mutex:        &sync.Mutex{},
		fileConsumer: fileConsumer,
		fileProducer: fileProducer,
	}
}

func (l *LinksStorage) AddLink(link models.Link) (models.Link, error) {
	l.addLinksToMap([]models.Link{link})

	err := l.writeFile(link)
	if err != nil {
		return link, err
	}

	return link, nil
}

func (l *LinksStorage) AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	l.addLinksToMap(links)

	for _, link := range links {
		err := l.writeFile(link)
		if err != nil {
			return links, err
		}
	}
	return links, nil
}

func (l *LinksStorage) GetLink(value string) (string, bool, error) {
	result, ok := l.linksMap[value]
	return result.OriginalURL, ok, nil
}

func (l *LinksStorage) addLinksToMap(links []models.Link) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, v := range links {
		l.linksMap[v.ShortURL] = v
	}
}

func (l *LinksStorage) InitStorage() error {
	rows, err := l.fileConsumer.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = models.Link{ShortURL: row.OriginalURL, OriginalURL: row.OriginalURL, CorrelationID: row.ID}
	}
	logger.GetLogger().Info("Link map initialized")
	return nil
}

func (l *LinksStorage) writeFile(link models.Link) error {
	event := &fileJob.Event{
		ID:          link.CorrelationID,
		ShortURL:    link.ShortURL,
		OriginalURL: link.OriginalURL,
	}
	err := l.fileProducer.WriteEvent(event)
	if err != nil {
		return errors.New("write events error")
	}
	return nil
}

func (l LinksStorage) Ping() error {
	return nil
}

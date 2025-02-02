package links

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksStorage interface {
	AddLink(link models.Links) error
	GetLink(value string) (string, bool, error)
	Ping() error
	AddLinkBatch(ctx context.Context, links []models.Links) error
}

type fileProducer interface {
	WriteEvent(event *fileJob.Event) error
}

type Link struct {
	linksStorage linksStorage
	fileProducer fileProducer
}

func NewLinkService(linksStorage linksStorage, fileProducer fileProducer) *Link {
	return &Link{
		linksStorage: linksStorage,
		fileProducer: fileProducer,
	}
}

func (l *Link) Get(shortLink string) (string, error) {
	v, ok, err := l.linksStorage.GetLink(shortLink)
	if err != nil {
		return "", err
	}
	if !ok {
		return v, errors.New("link not found")
	}
	return v, nil
}

func (l *Link) Add(long string) (string, error) {
	link := models.Links{
		ShortURL:    uuid.New().String(),
		OriginalURL: long,
	}
	err := l.linksStorage.AddLink(link)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return link.ShortURL, errors.New("error adding link")
	}

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    link.ShortURL,
		OriginalURL: long,
	}
	err = l.fileProducer.WriteEvent(event)
	if err != nil {
		return link.ShortURL, errors.New("write event error")
	}

	return link.ShortURL, nil
}

func (l *Link) AddBatch(links []models.Links) ([]models.Links, error) {
	for i := range links {
		links[i].ShortURL = uuid.New().String()
	}

	err := l.linksStorage.AddLinkBatch(context.Background(), links)
	if err != nil {
		logger.GetLogger().Error("add link batch error", zap.Error(err))
		return links, err
	}

	//запись в файл
	if !config.IsDatabaseExist {
		for _, link := range links {
			event := &fileJob.Event{
				ID:          link.CorrelationID,
				ShortURL:    link.ShortURL,
				OriginalURL: link.OriginalURL,
			}
			err = l.fileProducer.WriteEvent(event)
			if err != nil {
				return nil, errors.New("write events error")
			}
		}
	}

	return links, nil
}

func (l *Link) Ping() error {
	err := l.linksStorage.Ping()
	if err != nil {
		return err
	}
	return nil
}

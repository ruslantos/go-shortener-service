package links

import (
	"errors"

	"github.com/google/uuid"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	mid "github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksStorage interface {
	AddLink(raw models.Links) error
	GetLink(value string) (string, bool, error)
	Ping() error
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
	short := uuid.New().String()
	err := l.linksStorage.AddLink(models.Links{
		ShortURL:    short,
		OriginalURL: long,
	})
	if err != nil {
		mid.GetLogger().Error(err.Error())
		return short, errors.New("error adding link")
	}
	return short, nil
}

func (l *Link) Ping() error {
	err := l.linksStorage.Ping()
	if err != nil {
		return err
	}
	return nil
}

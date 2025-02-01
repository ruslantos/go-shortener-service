package links

import (
	"errors"

	"github.com/google/uuid"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
)

type linksStorage interface {
	AddLink(raw string) string
	GetLink(value string) (key string, ok bool)
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
	v, ok := l.linksStorage.GetLink(shortLink)
	if !ok {
		return v, errors.New("link not found")
	}
	return v, nil
}

func (l *Link) Add(long string) (string, error) {
	short := l.linksStorage.AddLink(long)

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    short,
		OriginalURL: long,
	}
	err := l.fileProducer.WriteEvent(event)
	if err != nil {
		return short, errors.New("write event error")
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

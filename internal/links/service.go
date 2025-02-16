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

type linksDBStorage interface {
	AddLink(link models.Link) (models.Link, error)
	GetLink(value string) (string, bool, error)
	Ping() error
	AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
}

type linksMapStorage interface {
	AddLink(link models.Link) (models.Link, error)
	GetLink(value string) (string, bool, error)
	AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
}

type fileProducer interface {
	WriteEvent(event *fileJob.Event) error
}

type LinkService struct {
	linksDBStorage  linksDBStorage
	linksMapStorage linksMapStorage
	fileProducer    fileProducer
}

func NewLinkService(linksDBStorage linksDBStorage, linksMapStorage linksMapStorage, fileProducer fileProducer) *LinkService {
	return &LinkService{
		linksDBStorage:  linksDBStorage,
		linksMapStorage: linksMapStorage,
		fileProducer:    fileProducer,
	}
}

func (l *LinkService) Get(shortLink string) (string, error) {
	switch {
	case config.IsDatabaseExist:
		v, ok, err := l.linksDBStorage.GetLink(shortLink)
		if err != nil {
			return "", err
		}
		if !ok {
			return v, errors.New("link not found")
		}
		return v, nil
	default:
		link, ok, err := l.linksMapStorage.GetLink(shortLink)
		if err != nil {
			return "", err
		}
		if !ok {
			return link, errors.New("link not found")
		}
		return link, nil
	}

}

func (l *LinkService) Add(long string) (string, error) {
	link := models.Link{
		ShortURL:    uuid.New().String(),
		OriginalURL: long,
	}

	switch {
	case config.IsDatabaseExist:
		savedLink, err := l.linksDBStorage.AddLink(link)
		if err != nil {
			return savedLink.ShortURL, err
		}
	default:
		savedLink, err := l.linksMapStorage.AddLink(link)
		if err != nil {
			return savedLink.ShortURL, err
		}

		err = l.writeFile(link)
		if err != nil {
			return link.ShortURL, err
		}
	}

	return link.ShortURL, nil
}

func (l *LinkService) AddBatch(links []models.Link) ([]models.Link, error) {
	for i := range links {
		links[i].ShortURL = uuid.New().String()
	}
	var linksSaved []models.Link
	var err error

	switch {
	case config.IsDatabaseExist:
		linksSaved, err = l.linksDBStorage.AddLinkBatch(context.Background(), links)
		if err != nil {
			logger.GetLogger().Error("add link batch error", zap.Error(err))
			return linksSaved, err
		}
	default:
		linksSaved, err = l.linksMapStorage.AddLinkBatch(context.Background(), links)
		if err != nil {
			logger.GetLogger().Error("add link batch error", zap.Error(err))
			return linksSaved, err
		}

		for _, link := range linksSaved {
			err = l.writeFile(link)
			if err != nil {
				return linksSaved, err
			}
		}
	}

	return linksSaved, nil
}

func (l *LinkService) Ping() error {
	return l.linksDBStorage.Ping()
}

func (l *LinkService) writeFile(link models.Link) error {
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

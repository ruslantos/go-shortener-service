package links

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksStorage interface {
	AddLink(link models.Link) (models.Link, error)
	GetLink(value string) (string, bool, error)
	Ping() error
	AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
}

//type linksMapStorage interface {
//	AddLink(link models.Link) (models.Link, error)
//	GetLink(value string) (string, bool, error)
//	AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
//}

type LinkService struct {
	linksStorage linksStorage
}

func NewLinkService(linksStorage linksStorage) *LinkService {
	return &LinkService{
		linksStorage: linksStorage,
	}
}

func (l *LinkService) Get(ctx context.Context, shortLink string) (string, error) {
	v, ok, err := l.linksStorage.GetLink(shortLink)
	if err != nil {
		return "", err
	}
	if !ok {
		return v, errors.New("link not found")
	}
	return v, nil
}

func (l *LinkService) Add(ctx context.Context, long string) (string, error) {
	link := models.Link{
		ShortURL:    uuid.New().String(),
		OriginalURL: long,
	}

	//switch {
	//case config.IsDatabaseExist:
	savedLink, err := l.linksStorage.AddLink(link)
	if err != nil {
		return savedLink.ShortURL, err
	}
	//default:
	//	savedLink, err := l.linksMapStorage.AddLink(link)
	//	if err != nil {
	//		return savedLink.ShortURL, err
	//	}
	//
	//	err = l.writeFile(link)
	//	if err != nil {
	//		return link.ShortURL, err
	//	}
	//}

	return link.ShortURL, nil
}

func (l *LinkService) AddBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	for i := range links {
		links[i].ShortURL = uuid.New().String()
	}
	var linksSaved []models.Link
	var err error

	//switch {
	//case config.IsDatabaseExist:
	linksSaved, err = l.linksStorage.AddLinkBatch(context.Background(), links)
	if err != nil {
		logger.GetLogger().Error("add link batch error", zap.Error(err))
		return linksSaved, err
	}
	//default:
	//	linksSaved, err = l.linksMapStorage.AddLinkBatch(context.Background(), links)
	//	if err != nil {
	//		logger.GetLogger().Error("add link batch error", zap.Error(err))
	//		return linksSaved, err
	//	}
	//
	//	for _, link := range linksSaved {
	//		err = l.writeFile(link)
	//		if err != nil {
	//			return linksSaved, err
	//		}
	//	}
	//}

	return linksSaved, nil
}

func (l *LinkService) Ping(ctx context.Context) error {
	return l.linksStorage.Ping()
}

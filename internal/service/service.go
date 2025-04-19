package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type LinksStorage interface {
	AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error)
	GetLink(ctx context.Context, value string) (models.Link, error)
	Ping(ctx context.Context) error
	AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error)
	GetUserLinks(ctx context.Context, userID string) ([]models.Link, error)
	DeleteUserURLs(ctx context.Context, urls []DeletedURLs) error
	InitStorage() error
}

type LinkService struct {
	linksStorage LinksStorage
	deleteChan   chan DeletedURLs
}

type Config struct {
	StorageType string
}

type DeletedURLs struct {
	URLs   string
	UserID string
}

func NewLinkService(linksStorage LinksStorage) *LinkService {
	return &LinkService{
		linksStorage: linksStorage,
		deleteChan:   make(chan DeletedURLs, 100),
	}
}

func (l *LinkService) Get(ctx context.Context, shortLink string) (string, error) {
	v, err := l.linksStorage.GetLink(ctx, shortLink)
	if err != nil {
		return "", err
	}
	if v.IsExist != nil && !*v.IsExist {
		return v.ShortURL, internal_errors.ErrURLNotFound
	}
	if v.IsDeleted {
		return "", internal_errors.ErrURLDeleted
	}
	return v.OriginalURL, nil
}

func (l *LinkService) Add(ctx context.Context, long string) (string, error) {
	userID := getUserIDFromContext(ctx)

	link := models.Link{
		ShortURL:    uuid.New().String(),
		OriginalURL: long,
	}

	savedLink, err := l.linksStorage.AddLink(ctx, link, userID)
	if err != nil {
		return savedLink.ShortURL, err
	}

	return link.ShortURL, nil
}

func (l *LinkService) AddBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	for i := range links {
		links[i].ShortURL = uuid.New().String()
	}
	var linksSaved []models.Link
	var err error
	userID := getUserIDFromContext(ctx)

	linksSaved, err = l.linksStorage.AddLinkBatch(ctx, links, userID)
	if err != nil {
		logger.GetLogger().Error("add link batch error", zap.Error(err))
		return linksSaved, err
	}

	return linksSaved, nil
}

func (l *LinkService) Ping(ctx context.Context) error {
	return l.linksStorage.Ping(ctx)
}

func (l *LinkService) GetUserUrls(ctx context.Context) ([]models.Link, error) {
	userID := getUserIDFromContext(ctx)

	v, err := l.linksStorage.GetUserLinks(ctx, userID)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (l *LinkService) StartDeleteWorker(ctx context.Context) {
	logger.GetLogger().Info("start delete worker")

	var buffer []DeletedURLs
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(buffer) > 0 {
				logger.GetLogger().Error("delete urls from db error: ctx.Done()")
			}
			return

		case data := <-l.deleteChan:
			buffer = append(buffer, data)
			if len(buffer) >= 10 {
				err := l.linksStorage.DeleteUserURLs(ctx, buffer)
				if err != nil {
					logger.GetLogger().Error("delete urls from db error", zap.Error(err))
				}
				buffer = buffer[:0]
				timer.Reset(10 * time.Second)
				logger.GetLogger().Info("timer expired, deleting urls from db")
			}

		case <-timer.C:
			if len(buffer) > 0 {
				logger.GetLogger().Info("timer expired, deleting urls from db")
				err := l.linksStorage.DeleteUserURLs(ctx, buffer)
				if err != nil {
					logger.GetLogger().Error("delete urls from db error", zap.Error(err))
				}
				buffer = buffer[:0]
			}
		}
	}
}

func (l *LinkService) ConsumeDeleteURLs(data DeletedURLs) {
	l.deleteChan <- data
}

func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

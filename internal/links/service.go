package links

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	errors2 "github.com/ruslantos/go-shortener-service/internal/errors"
	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksStorage interface {
	AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error)
	GetLink(ctx context.Context, value string) (models.Link, bool, error)
	Ping(ctx context.Context) error
	AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error)
	GetUserLinks(ctx context.Context, userID string) ([]models.Link, error)
}

type LinkService struct {
	linksStorage linksStorage
}

func NewLinkService(linksStorage linksStorage) *LinkService {
	return &LinkService{
		linksStorage: linksStorage,
	}
}

func (l *LinkService) Get(ctx context.Context, shortLink string) (string, error) {
	v, ok, err := l.linksStorage.GetLink(ctx, shortLink)
	if err != nil {
		return "", err
	}
	if !ok {
		return v.ShortURL, errors.New("link not found")
	}
	if v.IsDeleted != nil && *v.IsDeleted {
		return "", errors2.ErrURLDeleted
	}
	return v.ShortURL, nil
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

func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	logger.GetLogger().Info("get userID", zap.String("userID", userID))
	return userID
}

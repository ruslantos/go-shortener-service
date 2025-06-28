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

// LinksStorage определяет интерфейс для работы с хранилищем ссылок.
type LinksStorage interface {
	// AddLink добавляет новую ссылку в хранилище для указанного пользователя.
	AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error)
	// GetLink возвращает ссылку по её короткому идентификатору.
	GetLink(ctx context.Context, value string) (models.Link, error)
	// Ping проверяет соединение с хранилищем.
	Ping(ctx context.Context) error
	// AddLinkBatch добавляет пакет ссылок в хранилище для указанного пользователя.
	AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error)
	// GetUserLinks возвращает все ссылки для указанного пользователя.
	GetUserLinks(ctx context.Context, userID string) ([]models.Link, error)
	// DeleteUserURLs удаляет указанные ссылки для пользователя.
	DeleteUserURLs(ctx context.Context, urls []DeletedURLs) error
	// InitStorage инициализирует хранилище.
	InitStorage() error
	// Close закрывает хранилище.
	Close() error
	// CountUsers возвращает количество пользователей.
	CountURLs(ctx context.Context) int
	// CountUsers возвращает количество пользователей.
	CountUsers(ctx context.Context) int
}

// LinkService предоставляет сервис для работы с ссылками.
type LinkService struct {
	linksStorage LinksStorage
	deleteChan   chan DeletedURLs
}

// Config содержит конфигурационные параметры для сервиса.
type Config struct {
	StorageType string
}

// DeletedURLs представляет структуру для удаления ссылок.
type DeletedURLs struct {
	URLs   string
	UserID string
}

// NewLinkService создает новый экземпляр LinkService.
func NewLinkService(linksStorage LinksStorage) *LinkService {
	return &LinkService{
		linksStorage: linksStorage,
		deleteChan:   make(chan DeletedURLs, 100),
	}
}

// Get возвращает оригинальную ссылку по короткому идентификатору.
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

// Add добавляет новую ссылку в хранилище.
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

// AddBatch добавляет пакет ссылок в хранилище.
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

// Ping проверяет соединение с хранилищем.
func (l *LinkService) Ping(ctx context.Context) error {
	return l.linksStorage.Ping(ctx)
}

// GetUserUrls возвращает все ссылки для указанного пользователя.
func (l *LinkService) GetUserUrls(ctx context.Context) ([]models.Link, error) {
	userID := getUserIDFromContext(ctx)

	v, err := l.linksStorage.GetUserLinks(ctx, userID)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// StartDeleteWorker запускает воркер для удаления ссылок.
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

// ConsumeDeleteURLs добавляет ссылку в канал для удаления.
func (l *LinkService) ConsumeDeleteURLs(data DeletedURLs) {
	l.deleteChan <- data
}

// getUserIDFromContext извлекает userID из контекста.
func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetStats возвращает статистику по ссылкам и пользователям.
func (l *LinkService) GetStats(ctx context.Context) (urls int, users int, err error) {
	urls = l.linksStorage.CountURLs(ctx)
	users = l.linksStorage.CountUsers(ctx)

	return urls, users, nil
}

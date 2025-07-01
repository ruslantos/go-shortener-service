package mapstorage

import (
	"context"
	"errors"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ruslantos/go-shortener-service/internal/models"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// LinksStorage реализует хранилище ссылок с использованием встроенной карты.
type LinksStorage struct {
	linksMap map[string]models.Link
	mutex    *sync.Mutex
}

// NewMapStorage создает новый экземпляр LinksStorage.
func NewMapStorage() *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]models.Link),
		mutex:    &sync.Mutex{},
	}
}

// AddLink добавляет новую ссылку в хранилище.
func (l *LinksStorage) AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error) {
	l.addLinksToMap([]models.Link{link})

	return link, nil
}

// AddLinkBatch добавляет пакет ссылок в хранилище.
func (l *LinksStorage) AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error) {
	l.addLinksToMap(links)

	return links, nil
}

// GetLink возвращает ссылку по её короткому идентификатору.
func (l *LinksStorage) GetLink(ctx context.Context, value string) (models.Link, error) {
	result := l.linksMap[value]
	link := models.Link{OriginalURL: result.OriginalURL}
	return link, nil
}

// addLinksToMap добавляет ссылки в карту ссылок.
func (l *LinksStorage) addLinksToMap(links []models.Link) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, v := range links {
		l.linksMap[v.ShortURL] = v
	}
}

// InitStorage инициализирует хранилище (в данном случае не выполняет никаких действий).
func (l *LinksStorage) InitStorage() error {
	return nil
}

// Ping проверяет соединение с хранилищем (в данном случае всегда возвращает nil).
func (l LinksStorage) Ping(context.Context) error {
	return nil
}

// GetUserLinks возвращает все ссылки для указанного пользователя.
func (l *LinksStorage) GetUserLinks(ctx context.Context, userID string) ([]models.Link, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var userLinks []models.Link
	for _, link := range l.linksMap {
		if link.UserID == userID {
			userLinks = append(userLinks, link)
		}
	}

	return userLinks, nil
}

// DeleteUserURLs удаляет указанные ссылки для пользователя.
func (l *LinksStorage) DeleteUserURLs(ctx context.Context, urls []service.DeletedURLs) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, url := range urls {
		if link, exists := l.linksMap[url.URLs]; exists {
			link.IsDeleted = true
			l.linksMap[url.URLs] = link
		} else {
			return errors.New("url not found in storage")
		}
	}

	return nil
}

// Close закрывает соединение с хранилищем (в данном случае не выполняет никаких действий).
func (l *LinksStorage) Close() error {
	return nil
}

// CountURLs возвращает количество ссылок в хранилище.
func (l *LinksStorage) CountURLs(ctx context.Context) int {
	return 0
}

// CountUsers возвращает количество пользователей в хранилище.
func (l *LinksStorage) CountUsers(ctx context.Context) int {
	return 0
}

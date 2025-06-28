package filestorage

import (
	"context"
	"errors"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// FileConsumer определяет интерфейс для чтения событий из файла.
type FileConsumer interface {
	ReadEvents() ([]*fileJob.Event, error)
}

// FileProducer определяет интерфейс для записи событий в файл.
type FileProducer interface {
	WriteEvent(event *fileJob.Event) error
}

// LinksStorage реализует хранилище ссылок с использованием файлов.
type LinksStorage struct {
	linksMap     map[string]models.Link
	mutex        *sync.Mutex
	fileConsumer FileConsumer
	fileProducer FileProducer
}

// NewFileStorage создает новый экземпляр LinksStorage.
func NewFileStorage(fileConsumer FileConsumer, fileProducer FileProducer) *LinksStorage {
	return &LinksStorage{
		linksMap:     make(map[string]models.Link),
		mutex:        &sync.Mutex{},
		fileConsumer: fileConsumer,
		fileProducer: fileProducer,
	}
}

// AddLink добавляет новую ссылку в хранилище и записывает её в файл.
func (l *LinksStorage) AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error) {
	l.addLinksToMap([]models.Link{link})

	err := l.writeFile(link)
	if err != nil {
		return link, err
	}

	return link, nil
}

// AddLinkBatch добавляет пакет ссылок в хранилище и записывает их в файл.
func (l *LinksStorage) AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error) {
	l.addLinksToMap(links)

	for _, link := range links {
		err := l.writeFile(link)
		if err != nil {
			return links, err
		}
	}
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

// InitStorage инициализирует хранилище, загружая данные из файла.
func (l *LinksStorage) InitStorage() error {
	rows, err := l.fileConsumer.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = models.Link{ShortURL: row.OriginalURL, OriginalURL: row.OriginalURL, CorrelationID: row.ID}
	}
	logger.GetLogger().Info("Link file storage initialized")
	return nil
}

// writeFile записывает событие в файл.
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

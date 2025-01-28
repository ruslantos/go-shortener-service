package storage

import (
	"strconv"
	"sync"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	mid "github.com/ruslantos/go-shortener-service/internal/middleware"
)

type file interface {
	ReadEvents() ([]*fileJob.Event, error)
}

type LinksStorage struct {
	linksMap map[string]string
	mutex    *sync.Mutex
	file     file
}

func NewLinksStorage(file file) *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]string),
		mutex:    &sync.Mutex{},
		file:     file,
	}
}

func (l LinksStorage) AddLink(raw string) string {
	return l.getShortValue(raw)
}

func (l LinksStorage) GetLink(value string) (string, bool) {
	result, ok := l.linksMap[value]
	return result, ok
}

func (l LinksStorage) getShortValue(raw string) string {
	l.mutex.Lock()
	count := len(l.linksMap)
	short := strconv.Itoa(count + 1)
	l.linksMap[short] = raw
	l.mutex.Unlock()

	return short
}

func (l LinksStorage) InitLinkMap() error {
	rows, err := l.file.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = row.OriginalURL
	}
	mid.GetLogger().Info("Links map initialized")
	return nil
}

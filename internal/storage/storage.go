package storage

import (
	"strconv"
	"sync"
)

type LinksStorage struct {
	linksMap map[string]string
	mutex    *sync.Mutex
}

func NewLinksStorage() *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]string),
		mutex:    &sync.Mutex{},
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

package storage

import (
	"math/rand"
	"sync"
	"time"
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
	l.mutex.Lock()
	newShort, ok := l.linksMap[raw]
	if ok {
		l.mutex.Unlock()
		return newShort
	}
	short := generateRandomString(10)

	l.linksMap[raw] = short
	l.mutex.Unlock()
	return short
}

func (l LinksStorage) GetLink(value string) (key string, ok bool) {
	for k, v := range l.linksMap {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}

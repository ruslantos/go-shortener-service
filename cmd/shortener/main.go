package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LinksStorage struct {
	linksMap map[string]string
	mutex    *sync.Mutex
}

func (l LinksStorage) mainPage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		body, err := io.ReadAll(req.Body)
		if err != nil || body == nil || len(body) == 0 {
			http.Error(res, "Error reading body", http.StatusBadRequest)
			return
		}

		short := l.addLink(string(body))

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(short))
	case http.MethodGet:
		q := req.URL.Path
		if len(q) == 0 || q == "/" {
			http.Error(res, "Invalid path", http.StatusBadRequest)
			return
		}
		v, ok := l.getLink(strings.Replace(q, "/", "", 1))
		if !ok {
			http.Error(res, "Invalid path", http.StatusBadRequest)
		}

		res.Header().Set("Location", v)
		res.WriteHeader(http.StatusTemporaryRedirect)
	default:
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	l := LinksStorage{linksMap: make(map[string]string), mutex: &sync.Mutex{}}
	mux.HandleFunc("/", l.mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func (l LinksStorage) addLink(raw string) string {
	l.mutex.Lock()
	newShort, ok := l.linksMap[raw]
	if ok {
		return newShort
	}
	short := generateRandomString(10)

	l.linksMap[raw] = short
	l.mutex.Unlock()
	return short
}

func (l LinksStorage) getLink(value string) (key string, ok bool) {
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

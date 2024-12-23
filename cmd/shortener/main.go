package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func mainPage(res http.ResponseWriter, req *http.Request) {
	meta := "Header ===============\r\n"
	for k, v := range req.Header {
		meta += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	meta += "Query parameters ===============\r\n"
	for k, v := range req.URL.Query() {
		meta += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	fmt.Println(meta)

	switch req.Method {
	case http.MethodPost:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Error reading body", http.StatusBadRequest)
			return
		}

		m := addLink()
		v, ok := m[string(body)]
		if !ok {
			http.Error(res, "Error reading body", http.StatusBadRequest)
		}

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(v))
	case http.MethodGet:
		q := req.URL.Path
		if !strings.Contains(q, "EwHXdJfB") {
			http.Error(res, "Invalid path", http.StatusBadRequest)
			return
		}
		v, ok := getLink(strings.Replace(q, "/", "", 1))
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
	mux.HandleFunc("/", mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func addLink() map[string]string {
	m := make(map[string]string)
	m["https://practicum.yandex.ru/"] = "EwHXdJfB"
	return m
}
func getLink(value string) (key string, ok bool) {
	m := addLink()
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

package getlink

import (
	"net/http"
	"strings"
)

type linksStorage interface {
	AddLink(raw string) string
	GetLink(value string) (key string, ok bool)
}

type Handler struct {
	linksStorage linksStorage
}

func New(linksStorage linksStorage) *Handler {
	return &Handler{linksStorage: linksStorage}
}

func (h *Handler) Handle(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Path
	if len(q) == 0 || q == "/" {
		http.Error(res, "Invalid path", http.StatusBadRequest)
		return
	}
	v, ok := h.linksStorage.GetLink(strings.Replace(q, "/", "", 1))
	if !ok {
		http.Error(res, "Unknown link", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", v)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

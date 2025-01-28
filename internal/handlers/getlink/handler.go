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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Path

	v, ok := h.linksStorage.GetLink(strings.Replace(q, "/", "", 1))
	if !ok {
		http.Error(w, "Unknown link", http.StatusBadRequest)
		return
	}

	//w.Header().Set("Content-Type", "text/html")
	w.Header().Add("Location", v)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(""))
}

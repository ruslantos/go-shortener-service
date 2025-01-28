package postlink

import (
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
)

type linksStorage interface {
	AddLink(raw string) string
	GetLink(value string) (key string, ok bool)
}

type file interface {
	WriteEvent(event *fileJob.Event) error
}

type Handler struct {
	linksStorage linksStorage
	file         file
}

func New(linksStorage linksStorage, file file) *Handler {
	return &Handler{linksStorage: linksStorage, file: file}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	short := h.linksStorage.AddLink(string(body))

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    short,
		OriginalURL: string(body),
	}
	err := h.file.WriteEvent(event)
	if err != nil {
		http.Error(w, "Write event error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.FlagShortURL + short))
}

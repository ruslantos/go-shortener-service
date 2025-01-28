package shorten

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
)

type linksStorage interface {
	AddLink(raw string) string
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
	bodyRaw, _ := io.ReadAll(r.Body)
	if len(bodyRaw) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var body ShortenRequest
	err := json.Unmarshal(bodyRaw, &body)
	if err != nil {
		http.Error(w, "Unmarshalling error", http.StatusBadRequest)
		return
	}

	short := h.linksStorage.AddLink(body.URL)
	resp := ShortenResponse{Result: config.FlagShortURL + short}
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    short,
		OriginalURL: body.URL,
	}
	err = h.file.WriteEvent(event)
	if err != nil {
		http.Error(w, "Write event error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(result)
	return
}

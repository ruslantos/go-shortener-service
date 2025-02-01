package shorten

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/config"
)

type linksService interface {
	Add(long string) (string, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
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

	short, err := h.linksService.Add(body.URL)
	if err != nil {
		http.Error(w, "Write event error", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{Result: config.FlagShortURL + short}
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(result)
}

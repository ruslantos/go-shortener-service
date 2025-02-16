package shorten

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/config"
	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
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
	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Reading body error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if len(bodyRaw) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var body ShortenRequest
	err = json.Unmarshal(bodyRaw, &body)
	if err != nil {
		http.Error(w, "Unmarshalling error", http.StatusBadRequest)
		return
	}

	respStatus := http.StatusCreated
	short, err := h.linksService.Add(body.URL)
	if err != nil {
		if errors.Is(err, internal_errors.ErrURLAlreadyExists) {
			respStatus = http.StatusConflict
		} else {
			http.Error(w, "Write event error", http.StatusInternalServerError)
			return
		}
	}

	resp := ShortenResponse{Result: config.FlagShortURL + short}
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	w.Write(result)
}

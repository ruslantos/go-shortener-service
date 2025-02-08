package postlink

import (
	"fmt"
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
	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	respStatus := http.StatusCreated
	short, err := h.linksService.Add(string(body))
	if err != nil {
		if internal_errors.IsClientError(err) {
			respStatus = http.StatusConflict
		} else {
			http.Error(w, fmt.Sprintf("add short link error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	w.Write([]byte(config.FlagShortURL + short))
}

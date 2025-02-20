package ping

import (
	"context"
	"fmt"
	"net/http"
)

type linksService interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	err := h.linksService.Ping(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to ping: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

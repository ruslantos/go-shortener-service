package ping

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
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
		logger.GetLogger().Error("filed to get ping", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to ping: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

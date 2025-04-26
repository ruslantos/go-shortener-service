package ping

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

// linksService интерфейс для сервиса, который обрабатывает пинг.
type linksService interface {
	Ping(ctx context.Context) error
}

// Handler обработчик для пинга.
type Handler struct {
	linksService linksService
}

// New создаёт новый обработчик для пинга.
func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

// Handle обрабатывает HTTP-запрос для пинга.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	err := h.linksService.Ping(r.Context())
	if err != nil {
		logger.GetLogger().Error("failed to get ping", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to ping: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

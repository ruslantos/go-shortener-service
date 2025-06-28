package getstats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

// StatsResponse тип для ответа со статистикой.
type StatsResponse struct {
	URLs  int `json:"urls"`  // количество сокращённых URL в сервисе
	Users int `json:"users"` // количество пользователей в сервисе
}

// statsService интерфейс для сервиса, который предоставляет статистику.
type service interface {
	GetStats(ctx context.Context) (urls int, users int, err error)
}

// Handler обработчик для получения статистики.
type Handler struct {
	statsService service
}

// New создаёт новый обработчик для получения статистики.
func New(statsService service) *Handler {
	return &Handler{statsService: statsService}
}

// Handle обрабатывает HTTP-запрос для получения статистики сервиса.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	urls, users, err := h.statsService.GetStats(r.Context())
	if err != nil {
		logger.GetLogger().Error("failed to get stats", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get stats: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	resp := StatsResponse{
		URLs:  urls,
		Users: users,
	}

	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

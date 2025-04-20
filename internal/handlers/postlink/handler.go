package postlink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

// linksService определяет интерфейс для работы с ссылками.
type linksService interface {
	Add(ctx context.Context, long string) (string, error)
}

// Handler представляет обработчик HTTP-запросов для создания коротких ссылок.
type Handler struct {
	linksService linksService
}

// New создает новый экземпляр Handler с заданным linksService.
func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

// Handle обрабатывает HTTP-запрос для получения оригинальной ссылки по короткому идентификатору.
// В данном случае метод используется для создания новой короткой ссылки из переданной оригинальной ссылки.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Reading body error", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	respStatus := http.StatusCreated
	short, err := h.linksService.Add(r.Context(), string(body))
	if err != nil {
		if errors.Is(err, internal_errors.ErrURLAlreadyExists) {
			respStatus = http.StatusConflict
		} else {
			logger.GetLogger().Error("add short link error", zap.Error(err))
			http.Error(w, fmt.Sprintf("add short link error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	w.Write([]byte(config.FlagShortURL + short))
}

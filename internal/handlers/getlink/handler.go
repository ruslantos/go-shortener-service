package getlink

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type linksService interface {
	Get(ctx context.Context, shortLink string) (string, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

// Handle обрабатывает запросы для получения оригинальной ссылки по короткому идентификатору.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Path

	long, err := h.linksService.Get(r.Context(), strings.Replace(q, "/", "", 1))
	if err != nil {
		// ссылка удалена
		if errors.Is(err, internal_errors.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		// ссылка не найдена
		if errors.Is(err, internal_errors.ErrURLNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		logger.GetLogger().Error("failed to get original_url", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get original_url: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(""))
}

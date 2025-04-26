package getuserurls

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

// UserURLsResponse тип для ответа с пользовательскими URL.
type UserURLsResponse []UserURLs

// UserURLs структура для представления пользовательского URL.
type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// linksService интерфейс для сервиса, который обрабатывает получение пользовательских URL.
type linksService interface {
	GetUserUrls(ctx context.Context) ([]models.Link, error)
}

// Handler обработчик для получения пользовательских URL.
type Handler struct {
	linksService linksService
}

// New создаёт новый обработчик для получения пользовательских URL.
func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

// Handle обрабатывает HTTP-запрос для получения оригинальной ссылки по короткому идентификатору.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	urls, err := h.linksService.GetUserUrls(r.Context())
	if err != nil {
		logger.GetLogger().Error("failed to get user urls", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get user urls: %s", err.Error()), http.StatusBadRequest)
		return
	}
	resp := prepareResponse(urls)
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	respStatus := http.StatusOK
	if len(resp) == 0 {
		respStatus = http.StatusNoContent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	w.Write(result)
}

// prepareResponse преобразует срез ссылок в формат ответа.
func prepareResponse(links []models.Link) UserURLsResponse {
	resp := UserURLsResponse{}
	for _, link := range links {
		resp = append(resp, UserURLs{
			ShortURL:    config.FlagShortURL + link.ShortURL,
			OriginalURL: link.OriginalURL,
		})
	}
	return resp
}

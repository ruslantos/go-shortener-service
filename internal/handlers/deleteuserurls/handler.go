package deleteuserurls

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// linkService интерфейс для сервиса, который обрабатывает удаление URL.
type linkService interface {
	ConsumeDeleteURLs(data service.DeletedURLs)
}

// Handler обработчик для удаления пользовательских URL.
type Handler struct {
	service linkService
}

// DeleteUserURLsRequest структура запроса на удаление пользовательских URL.
type DeleteUserURLsRequest []string

// New создаёт новый обработчик для удаления пользовательских URL.
func New(service linkService) *Handler {
	return &Handler{service: service}
}

// Handle обрабатывает запрос на удаление пользовательских URL.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Reading body error", http.StatusBadRequest)
		return
	}

	if len(bodyRaw) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var body DeleteUserURLsRequest
	err = json.Unmarshal(bodyRaw, &body)
	if err != nil {
		http.Error(w, "Unmarshalling error", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	for _, url := range body {
		urls := service.DeletedURLs{
			URLs:   url,
			UserID: userID,
		}
		h.service.ConsumeDeleteURLs(urls)
	}

	w.WriteHeader(http.StatusAccepted)
}

// getUserIDFromContext извлекает идентификатор пользователя из контекста запроса.
func getUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return "", errors.New("user not found from context")
	}
	return userID, nil
}

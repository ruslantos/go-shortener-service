package deleteuserurls

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

type linkService interface {
	ConsumeDeleteURLs(data service.DeletedURLs)
}

type Handler struct {
	service linkService
}

func New(service linkService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Reading body error", http.StatusBadRequest)
		return
	}

	if len(bodyRaw) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var body DeleteUserURLsResponse
	err = json.Unmarshal(bodyRaw, &body)
	if err != nil {
		http.Error(w, "Unmarshalling error", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r.Context())
	urls := service.DeletedURLs{
		URLs:   body,
		UserID: userID,
	}
	h.service.ConsumeDeleteURLs(urls)

	w.WriteHeader(http.StatusAccepted)
}

func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

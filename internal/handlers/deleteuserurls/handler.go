package deleteuserurls

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type queue interface {
	DeleteUserUrls(ctx context.Context, ids []string) error
}

type Handler struct {
	queue queue
}

func New(queue queue) *Handler {
	return &Handler{queue: queue}
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

	err = h.queue.DeleteUserUrls(r.Context(), body)
	if err != nil {
		logger.GetLogger().Error("delete user urls error", zap.Error(err))
	}
	w.WriteHeader(http.StatusAccepted)
	//w.Header().Set("Content-Type", "application/json")
	//w.Write(result)
}

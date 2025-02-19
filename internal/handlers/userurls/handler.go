package userurls

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/middleware/cookie"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksService interface {
	GetUserUrls(ctx context.Context) ([]models.Link, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(cookie.UserIDKey).(string)
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	urls, err := h.linksService.GetUserUrls(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("filed to get user urls: %s", err.Error()), http.StatusBadRequest)
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

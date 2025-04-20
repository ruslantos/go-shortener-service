package shortenbatch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksService interface {
	AddBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

// Handle обрабатывает HTTP-запрос для получения оригинальной ссылки по короткому идентификатору.
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

	var body ShortenBatchRequest
	err = json.Unmarshal(bodyRaw, &body)
	if err != nil || body == nil || len(body) == 0 {
		http.Error(w, "Unmarshalling body error", http.StatusBadRequest)
		return
	}

	links, err := h.linksService.AddBatch(r.Context(), prepareRequest(body))
	respStatus := http.StatusCreated
	if err != nil {
		if errors.Is(err, internal_errors.ErrURLAlreadyExists) {
			respStatus = http.StatusConflict
		} else {
			logger.GetLogger().Error("add batch shorten error", zap.Error(err))
			http.Error(w, "add batch shorten error", http.StatusInternalServerError)
			return
		}
	}

	resp := prepareResponse(links)
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respStatus)
	w.Write(result)
}

func prepareRequest(body ShortenBatchRequest) []models.Link {
	links := make([]models.Link, len(body))
	for i, link := range body {
		links[i] = models.Link{OriginalURL: link.OriginalURL, CorrelationID: link.CorrelationID}
	}
	return links
}

func prepareResponse(links []models.Link) ShortenBatchResponse {
	resp := ShortenBatchResponse{}
	for _, link := range links {
		resp = append(resp, BatchShortURLs{CorrelationID: link.CorrelationID, ShortURL: config.FlagShortURL + link.ShortURL})
	}
	return resp
}

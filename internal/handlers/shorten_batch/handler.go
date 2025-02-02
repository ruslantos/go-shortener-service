package shorten_batch

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type linksService interface {
	AddBatch(links []models.Links) ([]models.Links, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	bodyRaw, _ := io.ReadAll(r.Body)
	if len(bodyRaw) == 0 {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var body ShortenBatchRequest
	err := json.Unmarshal(bodyRaw, &body)
	if err != nil || body == nil || len(body) == 0 {
		http.Error(w, "Unmarshalling body error", http.StatusBadRequest)
		return
	}

	links, err := h.linksService.AddBatch(prepareRequest(body))
	if err != nil {
		http.Error(w, "add batch shorten error", http.StatusInternalServerError)
		return
	}

	resp := prepareResponse(links)
	result, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Marshalling error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(result)
}

func prepareRequest(body ShortenBatchRequest) []models.Links {
	links := make([]models.Links, len(body))
	for i, link := range body {
		links[i] = models.Links{OriginalURL: link.OriginalURL, CorrelationID: link.CorrelationID}
	}
	return links
}

func prepareResponse(links []models.Links) ShortenBatchResponse {
	resp := ShortenBatchResponse{}
	for _, link := range links {
		resp = append(resp, BatchShortURLs{CorrelationID: link.CorrelationID, ShortURL: config.FlagShortURL + link.ShortURL})
	}
	return resp
}

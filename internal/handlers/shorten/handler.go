package shorten

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ruslantos/go-shortener-service/internal/config"
)

type linksStorage interface {
	AddLink(raw string) string
}

type Handler struct {
	linksStorage linksStorage
}

func New(linksStorage linksStorage) *Handler {
	return &Handler{linksStorage: linksStorage}
}

func (h *Handler) Handle(c *gin.Context) {
	bodyRaw, _ := io.ReadAll(c.Request.Body)
	if len(bodyRaw) == 0 {
		c.Data(http.StatusBadRequest, "text/html", []byte("Error reading body"))
		return
	}

	var body ShortenRequest
	err := json.Unmarshal(bodyRaw, &body)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html", []byte("Unmarshalling error"))
		return
	}

	short := h.linksStorage.AddLink(body.Url)
	resp := ShortenResponse{Result: config.FlagShortURL + short}
	result, err := json.Marshal(resp)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html", []byte("Marshalling error"))
		return
	}

	c.Data(http.StatusCreated, "application/json", result)
}

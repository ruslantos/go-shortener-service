package postlink

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ruslantos/go-shortener-service/internal/config"
)

type linksStorage interface {
	AddLink(raw string) string
	GetLink(value string) (key string, ok bool)
}

type Handler struct {
	linksStorage linksStorage
}

func New(linksStorage linksStorage) *Handler {
	return &Handler{linksStorage: linksStorage}
}

func (h *Handler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || body == nil || len(body) == 0 || c.Request.Method != http.MethodPost {
		c.Data(http.StatusBadRequest, "text/html", []byte("Error reading body"))
		return
	}

	short := h.linksStorage.AddLink(string(body))

	c.Data(http.StatusCreated, "text/html", []byte(config.FlagShortUrl+short))
}

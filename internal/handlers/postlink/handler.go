package postlink

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/file"
)

type linksStorage interface {
	AddLink(raw string) string
	GetLink(value string) (key string, ok bool)
}

type file interface {
	WriteEvent(event *fileJob.Event) error
}

type Handler struct {
	linksStorage linksStorage
	file         file
}

func New(linksStorage linksStorage, file file) *Handler {
	return &Handler{linksStorage: linksStorage, file: file}
}

func (h *Handler) Handle(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	if len(body) == 0 {
		c.Data(http.StatusBadRequest, "text/html", []byte("Error reading body"))
		return
	}

	short := h.linksStorage.AddLink(string(body))

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    short,
		OriginalURL: string(body),
	}
	err := h.file.WriteEvent(event)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}

	c.Data(http.StatusCreated, "text/html", []byte(config.FlagShortURL+short))
}

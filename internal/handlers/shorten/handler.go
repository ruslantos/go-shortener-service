package shorten

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/file"
)

type linksStorage interface {
	AddLink(raw string) string
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

	short := h.linksStorage.AddLink(body.URL)
	resp := ShortenResponse{Result: config.FlagShortURL + short}
	result, err := json.Marshal(resp)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html", []byte("Marshalling error"))
		return
	}

	event := &fileJob.Event{
		ID:          uuid.New().String(),
		ShortURL:    short,
		OriginalURL: body.URL,
	}
	err = h.file.WriteEvent(event)
	if err != nil {
		c.Data(http.StatusInternalServerError, "application/json", []byte(err.Error()))
		return
	}

	c.Data(http.StatusCreated, "application/json", result)
}

package getlink

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	q := c.Request.URL.Path
	if len(q) == 0 || q == "/" || c.Request.Method != http.MethodGet {
		c.Data(http.StatusBadRequest, "text/html", []byte("Bad Request"))
		return
	}
	v, ok := h.linksStorage.GetLink(strings.Replace(q, "/", "", 1))
	if !ok {
		c.Data(http.StatusBadRequest, "text/html", []byte("Unknown link"))
		return
	}

	//res.Header().Set("Location", v)
	//res.WriteHeader(http.StatusTemporaryRedirect)
	c.Writer.Header().Set("Location", v)
	c.Data(http.StatusTemporaryRedirect, "text/html", nil)
}

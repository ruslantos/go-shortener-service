package postlink

import (
	"io"
	"net/http"
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

func (h *Handler) Handle(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil || body == nil || len(body) == 0 || req.Method != http.MethodPost {
		http.Error(res, "Error reading body", http.StatusBadRequest)
		return
	}

	short := h.linksStorage.AddLink(string(body))

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + short))
}

package getlink

import (
	"fmt"
	"net/http"
	"strings"
)

type linksService interface {
	Get(shortLink string) (string, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Path

	long, err := h.linksService.Get(strings.Replace(q, "/", "", 1))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get long ling: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(""))
}

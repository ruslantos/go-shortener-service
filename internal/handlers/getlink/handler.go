package getlink

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type linksService interface {
	Get(ctx context.Context, shortLink string) (string, error)
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Path

	long, err := h.linksService.Get(r.Context(), strings.Replace(q, "/", "", 1))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get long link: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(""))
}

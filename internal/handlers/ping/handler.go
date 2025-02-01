package ping

import (
	"fmt"
	"net/http"
)

type linksService interface {
	Ping() error
}

type Handler struct {
	linksService linksService
}

func New(linksService linksService) *Handler {
	return &Handler{linksService: linksService}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	err := h.linksService.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get long ling: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hoyci/bookday/pkg/httputil"
)

type Handler struct {
	service Service
}

func NewHTTPHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/drivers/status", h.getDriversStatus)
}

func (h *Handler) getDriversStatus(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.service.GetDriversStatus(r.Context())
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, statuses)
}

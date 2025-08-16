package order

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/httputil"
)

type Handler struct {
	service Service
}

func NewHTTPHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/orders", h.CreateOrder)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var dto CreateOrderDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid request body", fault.WithHTTPCode(http.StatusBadRequest)))
		return
	}

	order, err := h.service.CreateOrder(r.Context(), dto)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusCreated, order)
}

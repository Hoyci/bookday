package auth

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

func (h *Handler) RegisterPublicRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
	})
}

func (h *Handler) RegisterAdminRoutes(router chi.Router) {
	router.Route("/admin", func(r chi.Router) {
		r.Post("/users", h.createUserByAdmin)
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var dto RegisterUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid body", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err)))
		return
	}

	userResponse, err := h.service.Register(r.Context(), dto)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusCreated, userResponse)
}

func (h *Handler) createUserByAdmin(w http.ResponseWriter, r *http.Request) {
	var dto RegisterUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid body", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err)))
		return
	}

	userResponse, err := h.service.CreateUserByAdmin(r.Context(), dto)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusCreated, userResponse)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var dto LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid body", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err)))
		return
	}

	loginResponse, err := h.service.Login(r.Context(), dto)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, loginResponse)
}

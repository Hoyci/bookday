package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/httputil"
)

// Handler encapsula a dependência do serviço de autenticação.
type Handler struct {
	service Service
}

// NewHTTPHandler cria uma nova instância do handler de autenticação.
func NewHTTPHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes adiciona as rotas de autenticação ao router principal.
func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
	})
}

// register trata as requisições de registo de novos utilizadores.
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

// login trata as requisições de autenticação de utilizadores.
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

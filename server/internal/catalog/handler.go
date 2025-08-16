package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	fault "github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/httputil"
)

type Handler struct {
	service Service
}

func NewHTTPHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/books", h.CreateBook)
	router.Get("/books", h.ListBooks)
	router.Get("/books/{id}", h.GetBookByID)
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var dto CreateBookDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid request body", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err)))
		return
	}

	book, err := h.service.CreateBook(r.Context(), dto)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusCreated, book)
}

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.ListAllBooks(r.Context())
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, books)
}

func (h *Handler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.RespondWithError(w, fault.New("book id is required", fault.WithHTTPCode(http.StatusBadRequest)))
		return
	}

	book, err := h.service.GetBookDetails(r.Context(), id)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, book)
}

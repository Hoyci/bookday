package routing

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hoyci/bookday/internal/middleware"
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
	router.Post("/route/associate", h.associateDriver)
	router.Get("/route/current", h.getCurrentRoute)
	router.Patch("/stops/{id}", h.updateStopStatus)
}

func (h *Handler) associateDriver(w http.ResponseWriter, r *http.Request) {
	driverID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || driverID == "" {
		httputil.RespondWithError(w, fault.New("driver ID not found in context", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized)))
		return
	}

	route, err := h.service.AssociateDriverToRoute(r.Context(), driverID)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	response := RouteAssociationResponseDTO{
		ID:       route.ID(),
		Status:   string(route.Status()),
		DriverID: *route.driverID,
	}

	httputil.RespondWithJSON(w, http.StatusOK, response)
}

func (h *Handler) getCurrentRoute(w http.ResponseWriter, r *http.Request) {
	driverID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || driverID == "" {
		httputil.RespondWithError(w, fault.New("driver ID not found in context", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized)))
		return
	}

	route, err := h.service.GetActiveRouteForDriver(r.Context(), driverID)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	response := toRouteDetailDTO(route)
	httputil.RespondWithJSON(w, http.StatusOK, response)
}

func toRouteDetailDTO(route *DeliveryRoute) *RouteDetailDTO {
	stopDTOs := make([]RouteStopDTO, len(route.Stops()))
	for i, stop := range route.Stops() {
		orderDTOs := make([]OrderSummaryDTO, len(stop.OrderIDs()))
		for j, orderID := range stop.OrderIDs() {
			orderDTOs[j] = OrderSummaryDTO{ID: orderID}
		}

		stopDTOs[i] = RouteStopDTO{
			ID:        stop.id,
			Sequence:  stop.sequence,
			Address:   stop.address,
			Status:    string(stop.status),
			Latitude:  stop.latitude,
			Longitude: stop.longitude,
			Orders:    orderDTOs,
		}
	}

	return &RouteDetailDTO{
		ID:        route.id,
		Status:    string(route.status),
		UpdatedAt: route.updatedAt,
		Stops:     stopDTOs,
	}
}

func (h *Handler) updateStopStatus(w http.ResponseWriter, r *http.Request) {
	driverID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		httputil.RespondWithError(w, fault.New("driver ID not found in context", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized)))
		return
	}

	stopID := chi.URLParam(r, "id")
	if stopID == "" {
		httputil.RespondWithError(w, fault.New("stop id is required", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest)))
		return
	}

	var dto UpdateStopStatusDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		httputil.RespondWithError(w, fault.New("invalid request body", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err)))
		return
	}

	err := h.service.UpdateStopStatus(r.Context(), driverID, stopID, dto.Status)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	httputil.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Stop status updated successfully"})
}

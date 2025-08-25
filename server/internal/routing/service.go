package routing

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/internal/order"
	"github.com/hoyci/bookday/pkg/fault"
)

type deliveryPoint struct {
	Address   string
	OrderIDs  []string
	Latitude  float64
	Longitude float64
}

type service struct {
	routingRepo Repository
	orderRepo   order.Repository
	geocoder    Geocoder
	log         *log.Logger
}

func NewService(
	routingRepo Repository,
	orderRepo order.Repository,
	geocoder Geocoder,
	logger *log.Logger,
) Service {
	return &service{
		routingRepo: routingRepo,
		orderRepo:   orderRepo,
		geocoder:    geocoder,
		log:         logger,
	}
}

func (s *service) GenerateRoutes(ctx context.Context, cutoffTime time.Time) error {
	s.log.Info("starting daily route generation process")

	pendingOrders, err := s.orderRepo.FindPendingOrdersBefore(ctx, cutoffTime)
	if err != nil {
		s.log.Error("failed to fetch pending orders", "error", err)
		return err
	}
	if len(pendingOrders) == 0 {
		s.log.Info("no pending orders to route today")
		return nil
	}

	ordersByAddress := make(map[string][]string)
	for _, o := range pendingOrders {
		address := o.CustomerAddress()
		ordersByAddress[address] = append(ordersByAddress[address], o.ID())
	}

	s.log.Info("geocoding unique addresses...", "count", len(ordersByAddress))
	var deliveryPoints []deliveryPoint
	for address, orderIDs := range ordersByAddress {
		lat, lon, err := s.geocoder.Geocode(ctx, address)
		if err != nil {
			s.log.Warn("failed to geocode address, skipping orders for this address", "address", address, "error", err)
			continue
		}
		deliveryPoints = append(deliveryPoints, deliveryPoint{
			Address:   address,
			OrderIDs:  orderIDs,
			Latitude:  lat,
			Longitude: lon,
		})
	}

	if len(deliveryPoints) == 0 {
		s.log.Warn("no addresses could be geocoded, stopping route generation")
		return nil
	}

	s.log.Info("clustering delivery points into routes...", "point_count", len(deliveryPoints))
	routeClusters, err := clusterStops(deliveryPoints)
	if err != nil {
		s.log.Error("failed to cluster delivery points", "error", err)
		return err
	}

	var routesToSave []*DeliveryRoute

	s.log.Info("optimizing each route using TSP algorithm...", "cluster_count", len(routeClusters))
	for _, cluster := range routeClusters {
		if len(cluster) == 0 {
			continue
		}

		orderedPoints := optimizeRoute(cluster)

		routeID := uuid.NewString()
		var routeStops []*RouteStop
		for i, point := range orderedPoints {
			newStop, _ := NewRouteStop(uuid.NewString(), routeID, i+1, point.Address, point.Latitude, point.Longitude, point.OrderIDs)
			routeStops = append(routeStops, newStop)
		}
		newRoute, _ := NewDeliveryRoute(routeID, routeStops)
		routesToSave = append(routesToSave, newRoute)
	}

	if len(routesToSave) > 0 {
		s.log.Info("saving generated routes to the database...", "route_count", len(routesToSave))
		if err := s.routingRepo.CreateRoutesInTx(ctx, routesToSave); err != nil {
			s.log.Error("failed to save routes", "error", err)
			return err
		}
	}

	s.log.Info("daily route generation completed successfully")
	return nil
}

func (s *service) AssociateDriverToRoute(ctx context.Context, driverID string) (*DeliveryRoute, error) {
	s.log.Info("attempting to associate driver to a route", "driver_id", driverID)

	isActive, err := s.routingRepo.IsDriverOnActiveRoute(ctx, driverID)
	if err != nil {
		s.log.Error("failed to check driver's active route status", "driver_id", driverID, "error", err)
		return nil, err
	}
	if isActive {
		s.log.Warn("driver already has an active route", "driver_id", driverID)
		return nil, fault.New("driver is already on an active route", fault.WithKind(fault.KindConflict), fault.WithHTTPCode(http.StatusConflict))
	}

	pendingRoute, err := s.routingRepo.FindPendingRoute(ctx)
	if err != nil {
		var f *fault.Error
		if errors.As(err, &f) && f.Kind == fault.KindNotFound {
			s.log.Info("no pending routes available for assignment")
			return nil, fault.New("no delivery routes available at the moment", fault.WithKind(fault.KindNotFound), fault.WithHTTPCode(http.StatusNotFound))
		}
		s.log.Error("failed to find a pending route", "error", err)
		return nil, err
	}

	err = s.routingRepo.AssignDriverToRoute(ctx, pendingRoute.ID(), driverID)
	if err != nil {
		s.log.Error("failed to assign driver to the route", "driver_id", driverID, "route_id", pendingRoute.ID(), "error", err)
		return nil, err
	}

	s.log.Info("driver successfully associated with route", "driver_id", driverID, "route_id", pendingRoute.ID())

	pendingRoute.status = models.RouteStatusInProgress
	driverIDStr := driverID
	pendingRoute.driverID = &driverIDStr

	return pendingRoute, nil
}

func (s *service) GetActiveRouteForDriver(ctx context.Context, driverID string) (*DeliveryRoute, error) {
	s.log.Info("fetching active route for driver", "driver_id", driverID)

	route, err := s.routingRepo.FindActiveRouteByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (s *service) UpdateStopStatus(ctx context.Context, driverID, stopID string, newStatus string) error {
	dto := UpdateStopStatusDTO{Status: newStatus}
	if err := dto.Validate(); err != nil {
		return fault.New("invalid status provided", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	route, err := s.routingRepo.FindRouteByStopID(ctx, stopID)
	if err != nil {
		return err
	}

	if route.driverID == nil || *route.driverID != driverID {
		s.log.Warn("authorization failed: driver does not own this route", "driver_id", driverID, "route_id", route.ID())
		return fault.New("you are not authorized to update this stop", fault.WithKind(fault.KindForbidden), fault.WithHTTPCode(http.StatusForbidden))
	}

	stopStatus := models.RouteStopStatus(newStatus)

	s.log.Info("updating stop status", "stop_id", stopID, "new_status", newStatus)
	if err := s.routingRepo.UpdateStopStatusInTx(ctx, stopID, stopStatus); err != nil {
		s.log.Error("failed to update stop status transactionally", "stop_id", stopID, "error", err)
		return fault.New("could not update stop status", fault.WithError(err))
	}

	if err := s.routingRepo.CheckAndCompleteRoute(ctx, route.ID()); err != nil {
		s.log.Error("failed to check and complete route after stop update", "route_id", route.ID(), "error", err)
	}

	return nil
}

package routing

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/hoyci/bookday/internal/order"
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

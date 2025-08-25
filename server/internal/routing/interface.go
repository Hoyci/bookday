package routing

import (
	"context"
	"time"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type Geocoder interface {
	Geocode(ctx context.Context, address string) (latitude, longitude float64, err error)
}

type Repository interface {
	CreateRoutesInTx(ctx context.Context, routes []*DeliveryRoute) error
	IsDriverOnActiveRoute(ctx context.Context, driverID string) (bool, error)
	FindPendingRoute(ctx context.Context) (*DeliveryRoute, error)
	AssignDriverToRoute(ctx context.Context, routeID, driverID string) error
	FindActiveRouteByDriverID(ctx context.Context, driverID string) (*DeliveryRoute, error)
	FindRouteByStopID(ctx context.Context, stopID string) (*DeliveryRoute, error)
	UpdateStopStatusInTx(ctx context.Context, stopID string, stopStatus models.RouteStopStatus) error
	CheckAndCompleteRoute(ctx context.Context, routeID string) error
}

type Service interface {
	GenerateRoutes(ctx context.Context, cutoffTime time.Time) error
	AssociateDriverToRoute(ctx context.Context, driverID string) (*DeliveryRoute, error)
	GetActiveRouteForDriver(ctx context.Context, driverID string) (*DeliveryRoute, error)
	UpdateStopStatus(ctx context.Context, driverID, stopID string, newStatus string) error
}

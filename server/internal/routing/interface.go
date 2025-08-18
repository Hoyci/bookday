package routing

import (
	"context"
	"time"
)

type Geocoder interface {
	Geocode(ctx context.Context, address string) (latitude, longitude float64, err error)
}

type Repository interface {
	CreateRoutesInTx(ctx context.Context, routes []*DeliveryRoute) error
}

type Service interface {
	GenerateRoutes(ctx context.Context, cutoffTime time.Time) error
}

package routing

import (
	"time"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type DeliveryRoute struct {
	id        string
	status    models.DeliveryRouteStatus
	driverID  *string
	createdAt time.Time
	updatedAt time.Time
	stops     []*RouteStop
}

type RouteStop struct {
	id        string
	routeID   string
	sequence  int
	address   string
	status    models.RouteStopStatus
	latitude  float64
	longitude float64
	notes     *string
	orderIDs  []string
	updatedAt time.Time
}

func NewDeliveryRoute(id string, stops []*RouteStop) (*DeliveryRoute, error) {
	route := &DeliveryRoute{
		id:        id,
		status:    models.RouteStatusPending,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
		stops:     stops,
	}
	return route, nil
}

func NewRouteStop(id, routeID string, sequence int, address string, lat, lon float64, orderIDs []string) (*RouteStop, error) {
	stop := &RouteStop{
		id:        id,
		routeID:   routeID,
		sequence:  sequence,
		address:   address,
		status:    models.StopStatusPending,
		latitude:  lat,
		longitude: lon,
		orderIDs:  orderIDs,
	}
	return stop, nil
}

func (dr *DeliveryRoute) ID() string                         { return dr.id }
func (dr *DeliveryRoute) Status() models.DeliveryRouteStatus { return dr.status }
func (dr *DeliveryRoute) Stops() []*RouteStop                { return dr.stops }

func (rs *RouteStop) ID() string                     { return rs.id }
func (rs *RouteStop) RouteID() string                { return rs.routeID }
func (rs *RouteStop) Sequence() int                  { return rs.sequence }
func (rs *RouteStop) Address() string                { return rs.address }
func (rs *RouteStop) Status() models.RouteStopStatus { return rs.status }
func (rs *RouteStop) Latitude() float64              { return rs.latitude }
func (rs *RouteStop) Longitude() float64             { return rs.longitude }
func (rs *RouteStop) OrderIDs() []string             { return rs.orderIDs }
func (rs *RouteStop) UpdatedAt() time.Time           { return rs.updatedAt }

package routing

import (
	"context"
	"errors"
	"fmt"

	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/pkg/fault"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateRoutesInTx(ctx context.Context, routes []*DeliveryRoute) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, route := range routes {
			routeModel := models.DeliveryRouteModel{
				ID:     route.ID(),
				Status: models.DeliveryRouteStatus(route.Status()),
			}
			if err := tx.Create(&routeModel).Error; err != nil {
				return err
			}

			var allOrderIDsInRoute []string

			for _, stop := range route.Stops() {
				stopModel := models.RouteStopModel{
					ID:        stop.ID(),
					RouteID:   route.ID(),
					Sequence:  stop.Sequence(),
					Address:   stop.Address(),
					Status:    models.RouteStopStatus(stop.Status()),
					Latitude:  stop.Latitude(),
					Longitude: stop.Longitude(),
				}
				if err := tx.Create(&stopModel).Error; err != nil {
					return err
				}

				for _, orderID := range stop.OrderIDs() {
					joinRecord := map[string]any{
						"route_stop_id": stopModel.ID,
						"order_id":      orderID,
					}
					if err := tx.Table("route_stop_orders").Create(joinRecord).Error; err != nil {
						return err
					}
				}

				allOrderIDsInRoute = append(allOrderIDsInRoute, stop.OrderIDs()...)
			}

			if len(allOrderIDsInRoute) > 0 {
				err := tx.Model(&models.OrderModel{}).
					Where("id IN ?", allOrderIDsInRoute).
					Update("status", models.StatusOutForDelivery).Error
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *gormRepository) IsDriverOnActiveRoute(ctx context.Context, driverID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.DeliveryRouteModel{}).
		Where("driver_id = ? AND status = ?", driverID, models.RouteStatusInProgress).
		Count(&count).Error
	if err != nil {
		return false, fault.New("failed to check for driver's active routes", fault.WithError(err))
	}
	return count > 0, nil
}

func (r *gormRepository) FindPendingRoute(ctx context.Context) (*DeliveryRoute, error) {
	var routeModel models.DeliveryRouteModel
	err := r.db.WithContext(ctx).
		Where("status = ?", models.RouteStatusPending).
		Order("created_at asc").
		First(&routeModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fault.New("no pending routes available", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find a pending route", fault.WithError(err))
	}

	return &DeliveryRoute{id: routeModel.ID}, nil
}

func (r *gormRepository) AssignDriverToRoute(ctx context.Context, routeID, driverID string) error {
	fmt.Println("routeID", routeID)
	fmt.Println("driverID", driverID)

	result := r.db.WithContext(ctx).Model(&models.DeliveryRouteModel{}).
		Where("id = ?", routeID).
		Updates(map[string]any{
			"driver_id": driverID,
			"status":    string(models.RouteStatusInProgress),
		})

	if result.Error != nil {
		return fault.New("failed to assign driver to route", fault.WithError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fault.New("route not found for assignment", fault.WithKind(fault.KindNotFound))
	}

	return nil
}

func (r *gormRepository) FindActiveRouteByDriverID(ctx context.Context, driverID string) (*DeliveryRoute, error) {
	var routeModel models.DeliveryRouteModel
	err := r.db.WithContext(ctx).
		Preload("Stops", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Stops.Orders").
		Where("driver_id = ? AND status = ?", driverID, models.RouteStatusInProgress).
		First(&routeModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fault.New("no active route found for this driver", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find active route", fault.WithError(err))
	}

	return toDeliveryRouteEntity(&routeModel), nil
}

func toDeliveryRouteEntity(model *models.DeliveryRouteModel) *DeliveryRoute {
	stops := make([]*RouteStop, len(model.Stops))
	for i, stopModel := range model.Stops {
		orderIDs := make([]string, len(stopModel.Orders))
		for j, orderModel := range stopModel.Orders {
			orderIDs[j] = orderModel.ID
		}
		stops[i] = &RouteStop{
			id:        stopModel.ID,
			routeID:   stopModel.RouteID,
			sequence:  stopModel.Sequence,
			address:   stopModel.Address,
			status:    stopModel.Status,
			latitude:  stopModel.Latitude,
			longitude: stopModel.Longitude,
			notes:     stopModel.Notes,
			updatedAt: stopModel.UpdatedAt,
			orderIDs:  orderIDs,
		}
	}

	route := &DeliveryRoute{
		id:        model.ID,
		status:    model.Status,
		driverID:  model.DriverID,
		createdAt: model.CreatedAt,
		updatedAt: model.UpdatedAt,
		stops:     stops,
	}
	return route
}

func (r *gormRepository) FindRouteByStopID(ctx context.Context, stopID string) (*DeliveryRoute, error) {
	var stopModel models.RouteStopModel
	if err := r.db.WithContext(ctx).First(&stopModel, "id = ?", stopID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fault.New("stop not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find stop", fault.WithError(err))
	}

	var routeModel models.DeliveryRouteModel
	if err := r.db.WithContext(ctx).First(&routeModel, "id = ?", stopModel.RouteID).Error; err != nil {
		return nil, fault.New("failed to find route associated with the stop", fault.WithError(err))
	}

	return toDeliveryRouteEntity(&routeModel), nil
}

func (r *gormRepository) UpdateStopStatusInTx(ctx context.Context, stopID string, stopStatus models.RouteStopStatus, orderStatus models.OrderStatus) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.RouteStopModel{}).Where("id = ?", stopID).Update("status", stopStatus).Error; err != nil {
			return err
		}

		var orderIDs []string
		if err := tx.Table("route_stop_orders").Where("route_stop_id = ?", stopID).Pluck("order_id", &orderIDs).Error; err != nil {
			return err
		}

		if len(orderIDs) > 0 {
			if err := tx.Model(&models.OrderModel{}).Where("id IN ?", orderIDs).Update("status", orderStatus).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormRepository) CheckAndCompleteRoute(ctx context.Context, routeID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pendingStopsCount int64

		err := tx.Model(&models.RouteStopModel{}).
			Where("route_id = ?", routeID).
			Not("status IN ?", []string{string(models.StopStatusDelivered), string(models.StopStatusFailed)}).
			Count(&pendingStopsCount).Error
		if err != nil {
			return fault.New("failed to check remaining stops", fault.WithError(err))
		}

		if pendingStopsCount == 0 {
			err := tx.Model(&models.DeliveryRouteModel{}).
				Where("id = ?", routeID).
				Update("status", models.RouteStatusCompleted).Error
			if err != nil {
				return fault.New("failed to update route status to completed", fault.WithError(err))
			}
		}

		return nil
	})
}

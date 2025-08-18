package routing

import (
	"context"

	models "github.com/hoyci/bookday/internal/infra/database/model"
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

				ordersToAssociate := make([]models.OrderModel, len(stop.OrderIDs()))
				for i, orderID := range stop.OrderIDs() {
					ordersToAssociate[i] = models.OrderModel{ID: orderID}
				}
				if err := tx.Model(&stopModel).Association("Orders").Append(ordersToAssociate); err != nil {
					return err
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

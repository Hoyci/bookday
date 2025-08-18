package order

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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

func (r *gormRepository) CreateOrderInTx(ctx context.Context, order *Order) error {
	orderModel := models.OrderModel{
		ID:              order.ID(),
		CustomerName:    order.CustomerName(),
		CustomerAddress: order.CustomerAddress(),
		Status:          models.OrderStatus(order.Status()),
		TotalPrice:      order.TotalPrice(),
		CreatedAt:       order.CreatedAt(),
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&orderModel).Error; err != nil {
			return err
		}

		for _, item := range order.Items() {
			orderItemModel := models.OrderItemModel{
				ID:           uuid.NewString(),
				OrderID:      order.ID(),
				BookID:       item.BookID(),
				Quantity:     item.Quantity(),
				PricePerUnit: item.PriceAtPurchase(),
			}
			if err := tx.Create(&orderItemModel).Error; err != nil {
				return err
			}

			ledgerTx := models.StockLedgerModel{
				ID:              uuid.NewString(),
				BookID:          item.BookID(),
				TransactionType: models.TransactionTypeOutbound,
				Quantity:        item.Quantity(),
				ReferenceID:     order.ID(),
			}
			if err := tx.Create(&ledgerTx).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormRepository) FindOrderByID(ctx context.Context, id string) (*Order, error) {
	var orderModel models.OrderModel
	result := r.db.WithContext(ctx).Preload("Items").First(&orderModel, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fault.New("order not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, result.Error
	}

	var orderItems []*OrderItem
	for _, itemModel := range orderModel.Items {
		item, _ := NewOrderItem(itemModel.ID, itemModel.OrderID, itemModel.BookID, itemModel.Quantity, itemModel.PricePerUnit)
		orderItems = append(orderItems, item)
	}

	order, _ := NewOrder(orderModel.ID, orderModel.CustomerName, orderModel.CustomerAddress, orderModel.TotalPrice, orderItems)

	return order, nil
}

func (r *gormRepository) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	result := r.db.WithContext(ctx).Model(&models.OrderModel{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fault.New("order not found for update", fault.WithKind(fault.KindNotFound))
	}
	return nil
}

func (r *gormRepository) FindPendingOrdersBefore(ctx context.Context, cutoffTime time.Time) ([]*Order, error) {
	var orderModels []*models.OrderModel
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ? AND created_at < ?", models.StatusAwaitingShipment, cutoffTime).
		Find(&orderModels)

	if result.Error != nil {
		return nil, result.Error
	}

	var orders []*Order
	for _, model := range orderModels {
		var orderItems []*OrderItem
		for _, itemModel := range model.Items {
			item, _ := NewOrderItem(itemModel.ID, itemModel.OrderID, itemModel.BookID, itemModel.Quantity, itemModel.PricePerUnit)
			orderItems = append(orderItems, item)
		}
		order, _ := NewOrder(model.ID, model.CustomerName, model.CustomerAddress, model.TotalPrice, orderItems)
		orders = append(orders, order)
	}

	return orders, nil
}

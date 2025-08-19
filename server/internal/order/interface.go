package order

import (
	"context"
	"errors"
	"time"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	CreateOrderInTx(ctx context.Context, order *Order) error
	FindOrderByID(ctx context.Context, id string) (*Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error
	FindPendingOrdersBefore(ctx context.Context, cutoffTime time.Time) ([]*Order, error)
}

type Service interface {
	CreateOrder(ctx context.Context, userID string, dto CreateOrderDTO) (*OrderDTO, error)
	GetOrderDetails(ctx context.Context, id string) (*OrderDTO, error)
}

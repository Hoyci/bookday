package order

import (
	"time"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type Order struct {
	id              string
	customerID      string
	customerAddress string
	status          models.OrderStatus
	totalPrice      float64
	createdAt       time.Time
	items           []*OrderItem
}

type OrderItem struct {
	id              string
	orderID         string
	bookID          string
	quantity        int
	priceAtPurchase float64
}

func NewOrder(id, customerID, customerAddress string, totalPrice float64, items []*OrderItem) (*Order, error) {
	order := &Order{
		id:              id,
		customerID:      customerID,
		customerAddress: customerAddress,
		status:          models.StatusAwaitingShipment,
		totalPrice:      totalPrice,
		createdAt:       time.Now().UTC(),
		items:           items,
	}
	return order, nil
}

func NewOrderItem(id, orderID, bookID string, quantity int, priceAtPurchase float64) (*OrderItem, error) {
	item := &OrderItem{
		id:              id,
		orderID:         orderID,
		bookID:          bookID,
		quantity:        quantity,
		priceAtPurchase: priceAtPurchase,
	}
	return item, nil
}

func (o *Order) ID() string                 { return o.id }
func (o *Order) CustomerID() string         { return o.customerID }
func (o *Order) CustomerAddress() string    { return o.customerAddress }
func (o *Order) Status() models.OrderStatus { return o.status }
func (o *Order) TotalPrice() float64        { return o.totalPrice }
func (o *Order) CreatedAt() time.Time       { return o.createdAt }
func (o *Order) Items() []*OrderItem        { return o.items }

func (oi *OrderItem) ID() string               { return oi.id }
func (oi *OrderItem) OrderID() string          { return oi.orderID }
func (oi *OrderItem) BookID() string           { return oi.bookID }
func (oi *OrderItem) Quantity() int            { return oi.quantity }
func (oi *OrderItem) PriceAtPurchase() float64 { return oi.priceAtPurchase }

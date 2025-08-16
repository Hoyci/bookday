package order

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

type OrderDTO struct {
	ID              string         `json:"id"`
	CustomerName    string         `json:"customer_name"`
	CustomerAddress string         `json:"customer_address"`
	Status          string         `json:"status"`
	TotalPrice      float64        `json:"total_price"`
	CreatedAt       time.Time      `json:"created_at"`
	Items           []OrderItemDTO `json:"items"`
}

type OrderItemDTO struct {
	BookID          string  `json:"book_id"`
	Quantity        int     `json:"quantity"`
	PriceAtPurchase float64 `json:"price_at_purchase"`
}

type CreateOrderDTO struct {
	CustomerName    string               `json:"customer_name"`
	CustomerAddress string               `json:"customer_address"`
	Items           []CreateOrderItemDTO `json:"items"`
}

type CreateOrderItemDTO struct {
	BookID   string `json:"book_id"`
	Quantity int    `json:"quantity"`
}

func (dto CreateOrderDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.CustomerName, v.Required, v.Length(3, 100)),
		v.Field(&dto.CustomerAddress, v.Required, v.Length(10, 255)),
		v.Field(&dto.Items, v.Required, v.Length(1, 0)), // Pelo menos 1 item
		v.Field(&dto.Items), // Valida cada item na lista
	)
}

func (dto CreateOrderItemDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.BookID, v.Required),
		v.Field(&dto.Quantity, v.Required, v.Min(1)),
	)
}

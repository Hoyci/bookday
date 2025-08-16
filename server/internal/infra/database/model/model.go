// Package model defines the GORM data structures that map to the database schema.
package models

import "time"

type StockLedgerTransactionType string

const (
	TransactionTypeInbound  StockLedgerTransactionType = "inbound"
	TransactionTypeOutbound StockLedgerTransactionType = "outbound"
)

type StockLedgerModel struct {
	ID              string `gorm:"type:uuid;primary_key"`
	BookID          string `gorm:"type:uuid"`
	TransactionType StockLedgerTransactionType
	Quantity        int
	ReferenceID     string `gorm:"type:uuid"`
	CreatedAt       time.Time
}

func (StockLedgerModel) TableName() string {
	return "stock_ledger"
}

type OrderStatus string

const (
	StatusAwaitingShipment OrderStatus = "awaiting_shipment"
	StatusOutForDelivery   OrderStatus = "out_for_delivery"
	StatusDelivered        OrderStatus = "delivered"
	StatusDeliveryFailed   OrderStatus = "delivery_failed"
)

type OrderModel struct {
	ID              string `gorm:"type:uuid;primary_key"`
	CustomerName    string
	CustomerAddress string
	Status          OrderStatus `gorm:"type:order_status"`
	TotalPrice      float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Items           []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string {
	return "orders"
}

type OrderItemModel struct {
	ID           string `gorm:"type:uuid;primary_key"`
	OrderID      string `gorm:"type:uuid"`
	BookID       string `gorm:"type:uuid"`
	Quantity     int
	PricePerUnit float64
}

func (OrderItemModel) TableName() string {
	return "order_items"
}

type BookModel struct {
	ID           string `gorm:"type:uuid;primary_key"`
	Title        string
	Author       string
	ISBN         string
	CatalogPrice float64
	CreatedAt    time.Time
}

func (BookModel) TableName() string {
	return "books"
}

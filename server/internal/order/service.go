package order

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/hoyci/bookday/internal/catalog"
	"github.com/hoyci/bookday/pkg/fault"
)

type service struct {
	orderRepo   Repository
	catalogRepo catalog.Repository // Dependência injetada
	log         *log.Logger
}

func NewService(orderRepo Repository, catalogRepo catalog.Repository, logger *log.Logger) Service {
	return &service{
		orderRepo:   orderRepo,
		catalogRepo: catalogRepo,
		log:         logger,
	}
}

func (s *service) CreateOrder(ctx context.Context, dto CreateOrderDTO) (*OrderDTO, error) {
	s.log.Info("starting order creation process", "customer", dto.CustomerName)
	if err := dto.Validate(); err != nil {
		return nil, fault.New("invalid order data", fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	var total float64
	var orderItems []*OrderItem
	booksToVerify := make(map[string]int) // Mapa para verificar o estoque total de cada livro

	for _, itemDTO := range dto.Items {
		booksToVerify[itemDTO.BookID] += itemDTO.Quantity
	}

	for bookID, quantity := range booksToVerify {
		stock, err := s.catalogRepo.GetAvailableStockCount(ctx, bookID)
		if err != nil {
			s.log.Error("failed to get stock for book", "book_id", bookID, "error", err)
			return nil, fault.New("failed to verify stock", fault.WithHTTPCode(http.StatusInternalServerError))
		}
		if stock < quantity {
			s.log.Warn("insufficient stock for order", "book_id", bookID, "needed", quantity, "available", stock)
			return nil, fault.New(fmt.Sprintf("insufficient stock for book %s", bookID), fault.WithHTTPCode(http.StatusConflict))
		}
	}

	for _, itemDTO := range dto.Items {
		book, err := s.catalogRepo.FindBookByID(ctx, itemDTO.BookID)
		if err != nil {
			return nil, err // O erro já vem formatado
		}
		price := book.CatalogPrice()
		total += price * float64(itemDTO.Quantity)
		item, _ := NewOrderItem(uuid.NewString(), "", book.ID(), itemDTO.Quantity, price)
		orderItems = append(orderItems, item)
	}

	order, _ := NewOrder(uuid.NewString(), dto.CustomerName, dto.CustomerAddress, total, orderItems)

	if err := s.orderRepo.CreateOrderInTx(ctx, order); err != nil {
		s.log.Error("failed to create order transaction", "error", err)
		return nil, fault.New("could not complete order", fault.WithHTTPCode(http.StatusInternalServerError))
	}

	s.log.Info("order created successfully", "order_id", order.ID())
	return toOrderDTO(order), nil
}

func (s *service) GetOrderDetails(ctx context.Context, id string) (*OrderDTO, error) {
	s.log.Info("getting order details", "order_id", id)
	order, err := s.orderRepo.FindOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fault.New("order not found", fault.WithHTTPCode(http.StatusNotFound), fault.WithKind(fault.KindNotFound))
		}
		s.log.Error("failed to find order by id", "order_id", id, "error", err)
		return nil, fault.New("unexpected database error", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}
	return toOrderDTO(order), nil
}

func toOrderDTO(order *Order) *OrderDTO {
	itemDTOs := make([]OrderItemDTO, 0, len(order.Items()))
	for _, item := range order.Items() {
		itemDTOs = append(itemDTOs, OrderItemDTO{
			BookID:          item.BookID(),
			Quantity:        item.Quantity(),
			PriceAtPurchase: item.PriceAtPurchase(),
		})
	}

	return &OrderDTO{
		ID:              order.ID(),
		CustomerName:    order.CustomerName(),
		CustomerAddress: order.CustomerAddress(),
		Status:          string(order.Status()),
		TotalPrice:      order.TotalPrice(),
		CreatedAt:       order.CreatedAt(),
		Items:           itemDTOs,
	}
}

package catalog

import (
	"context"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type Repository interface {
	FindBookByID(ctx context.Context, id string) (*Book, error)
	FindBookByISBN(ctx context.Context, isbn string) (*Book, error)
	FindAllBooks(ctx context.Context) ([]Book, error)

	CreateBookWithInitialLedger(ctx context.Context, book *Book, initialStock int) error
	AddLedgerTransaction(ctx context.Context, tx *models.StockLedgerModel) error
	GetAvailableStockCount(ctx context.Context, bookID string) (int, error)
}

type Service interface {
	ListAllBooks(ctx context.Context) ([]BookDTO, error)
	GetBookDetails(ctx context.Context, id string) (*BookDTO, error)
	CreateBook(ctx context.Context, dto CreateBookDTO) (*BookDTO, error)
}

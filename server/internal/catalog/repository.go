package catalog

import (
	"context"
	"errors"

	models "github.com/hoyci/bookday/internal/infra/database/model"
	fault "github.com/hoyci/bookday/pkg/fault"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateBookWithInitialLedger(ctx context.Context, book *Book, initialStock int) error {
	bookModel := models.BookModel{
		ID:           book.ID(),
		Title:        book.Title(),
		Author:       book.Author(),
		ISBN:         book.ISBN(),
		CatalogPrice: book.CatalogPrice(),
		CreatedAt:    book.CreatedAt(),
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&bookModel).Error; err != nil {
			return err
		}

		if initialStock > 0 {
			ledgerTx := models.StockLedgerModel{
				ID:              tx.Statement.Context.Value("transaction_id").(string),
				BookID:          book.ID(),
				TransactionType: "inbound",
				Quantity:        initialStock,
				ReferenceID:     book.ID(),
				CreatedAt:       book.CreatedAt(),
			}
			if err := tx.Create(&ledgerTx).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormRepository) AddLedgerTransaction(ctx context.Context, tx *models.StockLedgerModel) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *gormRepository) GetAvailableStockCount(ctx context.Context, bookID string) (int, error) {
	var inbound, outbound int

	result := r.db.WithContext(ctx).Model(&models.StockLedgerModel{}).
		Where("book_id = ? AND transaction_type = 'inbound'", bookID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&inbound)

	if result.Error != nil {
		return 0, fault.New("failed to query inbound stock", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	result = r.db.WithContext(ctx).Model(&models.StockLedgerModel{}).
		Where("book_id = ? AND transaction_type = 'outbound'", bookID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&outbound)

	if result.Error != nil {
		return 0, fault.New("failed to query outbound stock", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	return inbound - outbound, nil
}

func (r *gormRepository) FindBookByID(ctx context.Context, id string) (*Book, error) {
	var bookModel models.BookModel
	result := r.db.WithContext(ctx).First(&bookModel, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fault.New("book not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find book by ID", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	bookEntity, _ := NewBook(bookModel.ID, bookModel.Title, bookModel.Author, bookModel.ISBN, bookModel.CatalogPrice)

	stock, err := r.GetAvailableStockCount(ctx, bookModel.ID)
	if err != nil {
		return nil, err
	}
	bookEntity.SetAvailableStock(stock)

	return bookEntity, nil
}

func (r *gormRepository) FindBookByISBN(ctx context.Context, isbn string) (*Book, error) {
	var bookModel models.BookModel
	result := r.db.WithContext(ctx).First(&bookModel, "isbn = ?", isbn)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fault.New("book not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find book by ISBN", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	bookEntity, _ := NewBook(bookModel.ID, bookModel.Title, bookModel.Author, bookModel.ISBN, bookModel.CatalogPrice)

	stock, err := r.GetAvailableStockCount(ctx, bookModel.ID)
	if err != nil {
		return nil, err
	}
	bookEntity.SetAvailableStock(stock)

	return bookEntity, nil
}

func (r *gormRepository) FindAllBooks(ctx context.Context) ([]Book, error) {
	var bookModels []models.BookModel
	result := r.db.WithContext(ctx).
		Order("title asc").
		Find(&bookModels)

	if result.Error != nil {
		return nil, fault.New("failed to find all books", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	var books []Book
	for _, m := range bookModels {
		bookEntity, _ := NewBook(m.ID, m.Title, m.Author, m.ISBN, m.CatalogPrice)

		stock, err := r.GetAvailableStockCount(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		bookEntity.SetAvailableStock(stock)
		books = append(books, *bookEntity)
	}

	return books, nil
}

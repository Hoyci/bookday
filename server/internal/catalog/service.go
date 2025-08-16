package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/hoyci/bookday/pkg/fault"
)

type service struct {
	repo Repository
	log  *log.Logger
}

func NewService(repo Repository, logger *log.Logger) Service {
	return &service{
		repo: repo,
		log:  logger,
	}
}

func (s *service) CreateBook(ctx context.Context, dto CreateBookDTO) (*BookDTO, error) {
	s.log.Info("starting book creation process", "isbn", dto.ISBN, "title", dto.Title)

	if err := dto.Validate(); err != nil {
		s.log.Warn("validation failed for create book DTO", "error", err)
		return nil, fault.New(
			"invalid input for create book",
			fault.WithHTTPCode(http.StatusBadRequest),
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		)
	}

	existingBook, err := s.repo.FindBookByISBN(ctx, dto.ISBN)
	if err != nil {
		var f *fault.Error
		if !errors.As(err, &f) || f.Kind != fault.KindNotFound {
			s.log.Error("failed to check if book already exists", "error", err)
			return nil, fault.New("unexpected database error", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
		}
	}

	if existingBook != nil {
		s.log.Warn("attempted to create a book with an existing ISBN", "isbn", dto.ISBN)
		return nil, fault.New(
			fmt.Sprintf("book with ISBN %s already exists", dto.ISBN),
			fault.WithHTTPCode(http.StatusConflict),
			fault.WithKind(fault.KindConflict),
		)
	}

	bookID := uuid.NewString()
	book, err := NewBook(bookID, dto.Title, dto.Author, dto.ISBN, dto.CatalogPrice)
	if err != nil {
		s.log.Error("failed to create book entity", "error", err)
		return nil, err
	}

	transactionID := uuid.NewString()
	ctxWithTxID := context.WithValue(ctx, "transaction_id", transactionID)

	if err := s.repo.CreateBookWithInitialLedger(ctxWithTxID, book, dto.InitialStock); err != nil {
		s.log.Error("failed to persist book and initial stock ledger entry", "error", err)
		return nil, fault.New("failed to save book and initial stock", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	s.log.Info("book created successfully with initial stock ledger entry", "book_id", book.ID(), "initial_stock", dto.InitialStock)

	book.SetAvailableStock(dto.InitialStock)
	return toBookDTO(book), nil
}

func (s *service) ListAllBooks(ctx context.Context) ([]BookDTO, error) {
	s.log.Info("listing all books")
	books, err := s.repo.FindAllBooks(ctx)
	if err != nil {
		s.log.Error("failed to find all books", "error", err)
		return nil, fault.New("unexpected database error", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}
	var dtos []BookDTO
	for _, b := range books {
		dtos = append(dtos, *toBookDTO(&b))
	}
	return dtos, nil
}

func (s *service) GetBookDetails(ctx context.Context, id string) (*BookDTO, error) {
	s.log.Info("getting book details", "book_id", id)
	book, err := s.repo.FindBookByID(ctx, id)
	if err != nil {
		var f *fault.Error
		if errors.As(err, &f) && f.Kind == fault.KindNotFound {
			return nil, fault.New("book not found", fault.WithHTTPCode(http.StatusNotFound), fault.WithKind(fault.KindNotFound))
		}
		s.log.Error("failed to find book by id", "book_id", id, "error", err)
		return nil, fault.New("unexpected database error", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}
	return toBookDTO(book), nil
}

func toBookDTO(b *Book) *BookDTO {
	return &BookDTO{
		ID:             b.ID(),
		Title:          b.Title(),
		Author:         b.Author(),
		ISBN:           b.ISBN(),
		CatalogPrice:   b.CatalogPrice(),
		AvailableStock: b.AvailableStock(),
		CreatedAt:      b.CreatedAt(),
	}
}

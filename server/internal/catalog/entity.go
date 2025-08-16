package catalog

import (
	"net/http"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	fault "github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/validator"
)

type Book struct {
	id             string
	title          string
	author         string
	isbn           string
	catalogPrice   float64
	availableStock int
	createdAt      time.Time
}

func NewBook(id, title, author, isbn string, catalogPrice float64) (*Book, error) {
	b := &Book{
		id:           id,
		title:        title,
		author:       author,
		isbn:         isbn,
		catalogPrice: catalogPrice,
		createdAt:    time.Now().UTC(),
	}

	if err := b.validate(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Book) validate() error {
	err := v.ValidateStruct(b,
		v.Field(&b.title, v.Required.Error("title is required"), v.Length(1, 255)),
		v.Field(&b.author, v.Required.Error("author is required"), v.Length(1, 255)),
		v.Field(&b.isbn, v.Required.Error("isbn is required"), validator.IsISBN),
		v.Field(&b.catalogPrice, v.Required.Error("catalog price is required"), v.Min(0.01)),
	)
	if err != nil {
		return fault.New(
			"book entity validation failed",
			fault.WithHTTPCode(http.StatusUnprocessableEntity),
			fault.WithKind(fault.KindValidation),
			fault.WithError(err),
		)
	}
	return nil
}

func (b *Book) ID() string            { return b.id }
func (b *Book) Title() string         { return b.title }
func (b *Book) Author() string        { return b.author }
func (b *Book) ISBN() string          { return b.isbn }
func (b *Book) CatalogPrice() float64 { return b.catalogPrice }
func (b *Book) AvailableStock() int   { return b.availableStock }
func (b *Book) CreatedAt() time.Time  { return b.createdAt }

func (b *Book) SetAvailableStock(count int) {
	b.availableStock = count
}

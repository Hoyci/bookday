package catalog

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hoyci/bookday/pkg/validator"
)

type BookDTO struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Author         string    `json:"author"`
	ISBN           string    `json:"isbn"`
	CatalogPrice   float64   `json:"catalog_price"`
	AvailableStock int       `json:"available_stock"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateBookDTO struct {
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	ISBN         string  `json:"isbn"`
	CatalogPrice float64 `json:"catalog_price"`
	InitialStock int     `json:"initial_stock"`
}

func (dto CreateBookDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.Title, v.Required.Error("title is required"), v.Length(1, 255)),
		v.Field(&dto.Author, v.Required.Error("author is required"), v.Length(1, 255)),
		v.Field(&dto.ISBN, v.Required.Error("isbn is required"), validator.IsISBN),
		v.Field(&dto.CatalogPrice, v.Required.Error("catalog_price is required"), v.Min(0.01)),
		v.Field(&dto.InitialStock, v.Required.Error("initial_stock is required"), v.Min(1)),
	)
}

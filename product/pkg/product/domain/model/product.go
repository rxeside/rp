package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Price     float64
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ProductRepository interface {
	NextID() (uuid.UUID, error)
	Store(product *Product) error
	Find(id uuid.UUID) (*Product, error)
	Remove(id uuid.UUID) error

	ReserveStock(id uuid.UUID, quantity int) error
	ReleaseStock(id uuid.UUID, quantity int) error
}

package data

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus int

const (
	Open OrderStatus = iota
	Pending
	Paid
	Cancelled
)

type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type OrderItem struct {
	OrderID    uuid.UUID
	ProductID  uuid.UUID
	Count      int
	TotalPrice float64
}

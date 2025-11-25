package data

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus int

const (
	Pending PaymentStatus = iota
	Processing
	Succeeded
	Failed
	Cancelled
)

type Payment struct {
	ID        uuid.UUID
	WalletID  uuid.UUID
	OrderID   uuid.UUID
	Amount    float64
	Status    PaymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

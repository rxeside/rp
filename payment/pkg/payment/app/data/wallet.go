package data

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

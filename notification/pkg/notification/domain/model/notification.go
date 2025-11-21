package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OrderID   uuid.UUID
	Status    string
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Простой нотификейшен, но с копией юзера

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification *Notification) error
	Find(id uuid.UUID) (*Notification, error)
	Remove(id uuid.UUID) error
}

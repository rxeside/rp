package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationStatus string

const (
	StatusSuccess NotificationStatus = "success"
	StatusFailed  NotificationStatus = "failed"
)

type NotificationPayload struct {
	Email   string
	Message string
}

type Notification struct {
	ID         uuid.UUID
	Payload    NotificationPayload
	ExecutedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Status     *NotificationStatus
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification Notification) error
	Find(id uuid.UUID) (*Notification, error)
}

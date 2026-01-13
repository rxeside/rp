package data

import (
	"time"

	"github.com/google/uuid"

	"notification/pkg/notification/domain/model"
)

type NotificationStatus string

const (
	StatusSuccess NotificationStatus = "success"
	StatusFailed  NotificationStatus = "failed"
)

func NotificationStatusFromDomain(status *model.NotificationStatus) *NotificationStatus {
	if status == nil {
		return nil
	}
	result := NotificationStatus(*status)
	return &result
}

type NotificationPayload struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

type Notification struct {
	ID         uuid.UUID
	Payload    NotificationPayload
	ExecutedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Status     *NotificationStatus
}

package activity

import (
	"context"

	"github.com/google/uuid"

	appdata "notification/pkg/notification/app/data"
	"notification/pkg/notification/app/service"
)

func NewNotificationActivities(
	notificationService service.NotificationService,
) *NotificationActivities {
	return &NotificationActivities{
		notificationService: notificationService,
	}
}

type NotificationActivities struct {
	notificationService service.NotificationService
}

func (a *NotificationActivities) CreateNotification(ctx context.Context, payload appdata.NotificationPayload) (uuid.UUID, error) {
	return a.notificationService.CreateNotification(ctx, payload)
}

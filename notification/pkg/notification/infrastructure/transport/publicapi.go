package transport

import (
	"context"

	"notification/api/server/notificationinternal"
	"notification/pkg/notification/app/service"
)

func NewPaymentInternalAPI(
	notificationService service.NotificationService,
) notificationinternal.NotificationInternalServiceServer {
	return &notificationInternalAPI{
		notificationService: notificationService,
	}
}

type notificationInternalAPI struct {
	notificationService service.NotificationService

	notificationinternal.UnsafeNotificationInternalServiceServer
}

func (n notificationInternalAPI) Ping(_ context.Context, _ *notificationinternal.PingRequest) (*notificationinternal.PingResponse, error) {
	// TODO implement me
	panic("implement me")
}

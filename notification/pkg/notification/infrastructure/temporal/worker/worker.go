package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"notification/pkg/notification/app/service"
	"notification/pkg/notification/infrastructure/temporal"
	"notification/pkg/notification/infrastructure/temporal/activity"
	"notification/pkg/notification/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	notificationService service.NotificationService,
	userService service.UserService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewNotificationActivities(notificationService))
	w.RegisterActivity(activity.NewUserActivities(userService))
	w.RegisterWorkflow(workflows.CreateUserWorkflow)
	w.RegisterWorkflow(workflows.UserUpdatedWorkflow)
	return w
}

package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"user/pkg/user/application/service"
	"user/pkg/user/infrastructure/temporal"
	"user/pkg/user/infrastructure/temporal/activity"
	"user/pkg/user/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	userService service.UserService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewUserServiceActivities(userService))
	w.RegisterWorkflow(workflows.UserUpdatedWorkflow)
	return w
}

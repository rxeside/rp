package worker

import (
	// Rename local package to avoid collision with sdk activity
	appactivity "order/pkg/order/infrastructure/activity"
	"order/pkg/order/infrastructure/temporal/workflows"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"order/pkg/order/app/service"
	"order/pkg/order/infrastructure/temporal"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	os service.OrderService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})

	acts := appactivity.NewOrderActivities(os)

	w.RegisterActivityWithOptions(acts.SetOrderStatusActivity, activity.RegisterOptions{Name: "SetOrderStatusActivity"})

	w.RegisterWorkflow(workflows.CreateOrderSaga)
	return w
}

package worker

import (
	"order/pkg/order/infrastructure/activity"
	"order/pkg/order/infrastructure/temporal/workflows"

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

	w.RegisterActivity(activity.NewOrderActivities(os))

	w.RegisterWorkflow(workflows.CreateOrderSaga)
	return w
}

package worker

import (
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
	_ service.OrderService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	return w
}

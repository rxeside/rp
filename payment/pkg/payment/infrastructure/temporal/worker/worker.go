package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"payment/pkg/payment/app/service"
	"payment/pkg/payment/infrastructure/temporal"
	"payment/pkg/payment/infrastructure/temporal/activity"
	"payment/pkg/payment/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	paymentService service.PaymentService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
	w.RegisterActivity(activity.NewPaymentServiceActivities(paymentService))
	w.RegisterWorkflow(workflows.UserUpdatedWorkflow)
	return w
}

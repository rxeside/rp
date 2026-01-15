package worker

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"payment/pkg/payment/app/service"
	"payment/pkg/payment/infrastructure/temporal"
	appactivity "payment/pkg/payment/infrastructure/temporal/activity"
	"payment/pkg/payment/infrastructure/temporal/workflows"
)

func InterruptChannel() <-chan interface{} {
	return worker.InterruptCh()
}

func NewWorker(
	temporalClient client.Client,
	walletService service.WalletService,
) worker.Worker {
	w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})

	acts := appactivity.NewWalletServiceActivities(walletService)

	// Explicitly register activities with string names
	w.RegisterActivityWithOptions(acts.CreateWallet, activity.RegisterOptions{Name: "CreateWallet"})
	w.RegisterActivityWithOptions(acts.ChargeWallet, activity.RegisterOptions{Name: "ChargeWallet"})
	w.RegisterActivityWithOptions(acts.RefundWallet, activity.RegisterOptions{Name: "RefundWallet"})

	w.RegisterWorkflow(workflows.CreateWalletWorkflow)
	return w
}

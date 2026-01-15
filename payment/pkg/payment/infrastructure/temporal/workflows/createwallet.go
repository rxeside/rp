package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"payment/pkg/payment/domain/model"
)

func CreateWalletWorkflow(ctx workflow.Context, event model.UserCreated) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           "payment_task_queue",
		StartToCloseTimeout: time.Minute,
	})

	fmt.Println("CreateWalletWorkflow start")
	// CALL BY EXPLICIT STRING NAME "CreateWallet"
	return workflow.ExecuteActivity(ctx, "CreateWallet", event.UserID).Get(ctx, nil)
}

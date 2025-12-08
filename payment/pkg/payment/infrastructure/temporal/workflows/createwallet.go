package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/infrastructure/temporal/activity"
)

var walletActivities *activity.WalletServiceActivities

func CreateWalletWorkflow(ctx workflow.Context, event model.UserCreated) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	fmt.Println("CreateWalletWorkflow start")
	return workflow.ExecuteActivity(ctx, walletActivities.CreateWallet, event.UserID).Get(ctx, nil)
}

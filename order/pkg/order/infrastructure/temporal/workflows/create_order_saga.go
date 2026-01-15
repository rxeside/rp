package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type OrderSagaParams struct {
	OrderID    string
	UserID     string
	Items      []OrderItemParam
	TotalPrice float64
}

type OrderItemParam struct {
	ProductID string
	Quantity  int
}

func CreateOrderSaga(ctx workflow.Context, params OrderSagaParams) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Order Saga", "OrderID", params.OrderID)

	// 1. Reserve Products
	for _, item := range params.Items {
		ctxProduct := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			TaskQueue:           "product-task-queue",
			StartToCloseTimeout: time.Minute,
		})

		// CALL BY EXPLICIT STRING NAME "ReserveProduct"
		err := workflow.ExecuteActivity(ctxProduct, "ReserveProduct", item.ProductID, item.Quantity).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to reserve product", "Error", err)
			return setOrderStatus(ctx, params.OrderID, "Cancelled")
		}
	}

	// 2. Charge Wallet
	ctxPayment := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           "payment_task_queue",
		StartToCloseTimeout: time.Minute,
	})

	// CALL BY EXPLICIT STRING NAME "ChargeWallet"
	err := workflow.ExecuteActivity(ctxPayment, "ChargeWallet", params.UserID, params.TotalPrice).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to charge wallet", "Error", err)

		// Compensation: Release Products
		for _, item := range params.Items {
			ctxProduct := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{TaskQueue: "product-task-queue"})
			// CALL BY EXPLICIT STRING NAME "ReleaseProduct"
			_ = workflow.ExecuteActivity(ctxProduct, "ReleaseProduct", item.ProductID, item.Quantity).Get(ctx, nil)
		}

		return setOrderStatus(ctx, params.OrderID, "Cancelled")
	}

	// 3. Success
	return setOrderStatus(ctx, params.OrderID, "Paid")
}

func setOrderStatus(ctx workflow.Context, orderID, status string) error {
	// CALL BY EXPLICIT STRING NAME "SetOrderStatusActivity"
	// Note: using context from start of workflow (Order Task Queue)
	return workflow.ExecuteActivity(ctx, "SetOrderStatusActivity", orderID, status).Get(ctx, nil)
}

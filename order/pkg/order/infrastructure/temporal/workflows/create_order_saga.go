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

	for _, item := range params.Items {
		ctxProduct := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			TaskQueue:           "product-task-queue",
			StartToCloseTimeout: time.Minute,
		})

		err := workflow.ExecuteActivity(ctxProduct, "ReserveProduct", item.ProductID, item.Quantity).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to reserve product", "Error", err)
			return setOrderStatus(ctx, params.OrderID, "Cancelled")
		}
	}

	defer func() {
		// kompensacii nije
	}()

	ctxPayment := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           "payment_task_queue",
		StartToCloseTimeout: time.Minute,
	})

	err := workflow.ExecuteActivity(ctxPayment, "ChargeWallet", params.UserID, params.TotalPrice).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to charge wallet", "Error", err)

		for _, item := range params.Items {
			ctxProduct := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{TaskQueue: "product-task-queue"})
			_ = workflow.ExecuteActivity(ctxProduct, "ReleaseProduct", item.ProductID, item.Quantity).Get(ctx, nil)
		}

		return setOrderStatus(ctx, params.OrderID, "Cancelled")
	}

	return setOrderStatus(ctx, params.OrderID, "Paid")
}

func setOrderStatus(ctx workflow.Context, orderID, status string) error {
	return workflow.ExecuteActivity(ctx, "SetOrderStatusActivity", orderID, status).Get(ctx, nil)
}

package workflows

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"payment/pkg/payment/infrastructure/temporal/activity"
)

var paymentServiceActivities *activity.PaymentServiceActivities

func UserUpdatedWorkflow(_ workflow.Context) error {
	fmt.Println(paymentServiceActivities)

	return nil
}

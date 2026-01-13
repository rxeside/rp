package workflows

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"user/pkg/user/domain/model"
)

func UserUpdatedWorkflow(_ workflow.Context, event model.UserUpdated) error {
	fmt.Println("UserUpdatedWorkflow event = ", event)
	return nil
}

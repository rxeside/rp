package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"user/pkg/user/domain/model"
	"user/pkg/user/infrastructure/temporal/workflows"
)

const TaskQueue = "userservice_task_queue"

type WorkflowService interface {
	RunUserUpdatedWorkflow(ctx context.Context, id string, event model.UserUpdated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunUserUpdatedWorkflow(ctx context.Context, id string, event model.UserUpdated) error {
	fmt.Println("RunUserUpdatedWorkflow event = ", event)
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: TaskQueue,
		},
		workflows.UserUpdatedWorkflow, event,
	)
	return err
}

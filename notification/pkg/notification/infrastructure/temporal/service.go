package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/infrastructure/temporal/workflows"
)

const TaskQueue = "notification_task_queue"

type WorkflowService interface {
	RunCreateUserWorkflow(ctx context.Context, id string, event model.UserCreated) error
	RunUpdateUserWorkflow(ctx context.Context, id string, event model.UserUpdated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunCreateUserWorkflow(ctx context.Context, id string, event model.UserCreated) error {
	fmt.Println("RunCreateUserWorkflow event = ", event)
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: TaskQueue,
		},
		workflows.CreateUserWorkflow, event,
	)
	return err
}

func (s *workflowService) RunUpdateUserWorkflow(ctx context.Context, id string, event model.UserUpdated) error {
	fmt.Println("RunUpdateUserWorkflow event = ", event)
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

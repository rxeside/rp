package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/infrastructure/temporal/workflows"
)

const TaskQueue = "payment_task_queue"

type WorkflowService interface {
	RunCreateWalletWorkflow(ctx context.Context, id string, event model.UserCreated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunCreateWalletWorkflow(ctx context.Context, id string, event model.UserCreated) error {
	fmt.Println("RunCreateWalletWorkflow event = ", event)
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: TaskQueue,
		},
		workflows.CreateWalletWorkflow, event,
	)
	return err
}

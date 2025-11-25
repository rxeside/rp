package temporal

import (
	"context"

	"go.temporal.io/sdk/client"

	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/infrastructure/temporal/workflows"
)

const TaskQueue = "payment_task_queue"

type WorkflowService interface {
	RunUserUpdatedWorkflow(ctx context.Context, id string, event model.PaymentCreated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunUserUpdatedWorkflow(ctx context.Context, id string, event model.PaymentCreated) error {
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

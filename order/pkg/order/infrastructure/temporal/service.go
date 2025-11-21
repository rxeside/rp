package temporal

import (
	"go.temporal.io/sdk/client"
)

const TaskQueue = "order_task_queue"

type WorkflowService interface {
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

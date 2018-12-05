package archival

import (
	"context"
	"errors"
	"go.uber.org/cadence/.gen/go/shared"
)

// Client is used to store and retrieve blobs
type Client interface {
	Archive(
		ctx context.Context,
		domainName string,
		domainID string,
		workflowInfo *shared.WorkflowExecutionInfo,
		history *shared.History,
	) error

	GetWorkflowExecutionHistory(
		ctx context.Context,
		request *shared.GetWorkflowExecutionHistoryRequest,
	) (*shared.GetWorkflowExecutionHistoryResponse, error)

	ListClosedWorkflowExecutions(
		ctx context.Context,
		ListRequest *shared.ListClosedWorkflowExecutionsRequest,
	) (*shared.ListClosedWorkflowExecutionsResponse, error)
}

type nopClient struct{}

func (c *nopClient) Archive(
	ctx context.Context,
	domainName string,
	domainID string,
	workflowInfo *shared.WorkflowExecutionInfo,
	history *shared.History,
) error {
 	return errors.New("not implemented")
}

func (c *nopClient) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *shared.GetWorkflowExecutionHistoryRequest,
) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *nopClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	ListRequest *shared.ListClosedWorkflowExecutionsRequest,
) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	return nil, errors.New("not implemented")
}

func NewNopClient() Client {
	return &nopClient{}
}

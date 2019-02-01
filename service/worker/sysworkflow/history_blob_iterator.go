package sysworkflow

import (
	"context"
	"errors"
	"github.com/uber/cadence/client/public"
	"go.uber.org/cadence/.gen/go/shared"
)

var (
	errExecutionNotClosed = errors.New("cannot iterate over history, execution is not closed")
	errNoMoreItems = errors.New("iterator has no remaining elements")
)

type (
	HistoryBlobIterator interface {
		Next() (*HistoryBlob, error)
		HasNext() bool
	}

	historyBlobIterator struct {
		publicClient public.Client
		lastHistoryBlobReturned *HistoryBlob
		domain *string
		execution *shared.WorkflowExecution
		publicClientPageToken []byte
		failoverVersionLimit int64
		finished bool
		pageSizeLimit int
	}
)

func NewHistoryBlobIterator(
	publicClient public.Client,
	domain *string,
	execution *shared.WorkflowExecution,
	failoverVersionLimit int64,
	pageSizeLimit
) (HistoryBlobIterator, error) {
	itr := &historyBlobIterator{
		publicClient: publicClient,
		domain: domain,
		execution: execution,
		failoverVersionLimit: failoverVersionLimit,
	}
	closed, err := itr.executionClosed()
	if err != nil {
		return nil, err
	}
	if !closed {
		return nil, errExecutionNotClosed
	}
	return itr, nil
}

func (i *historyBlobIterator) Next() (*HistoryBlob, error) {
	if !i.HasNext() {
		return nil, errNoMoreItems
	}


	return nil, nil
}

func (i *historyBlobIterator) HasNext() bool {
	return !i.finished
}

func (i *historyBlobIterator) readCurrentPublicClientPage() (*shared.GetWorkflowExecutionHistoryResponse, error) {
	request := &shared.GetWorkflowExecutionHistoryRequest{
		Domain: i.domain,
		Execution: i.execution,
		NextPageToken: i.publicClientPageToken,
	}
	return i.publicClient.GetWorkflowExecutionHistory(context.Background(), request)
}

func (i *historyBlobIterator) executionClosed() (bool, error) {
	request := &shared.DescribeWorkflowExecutionRequest{
		Domain: i.domain,
		Execution: i.execution,
	}
	resp, err := i.publicClient.DescribeWorkflowExecution(context.Background(), request)
	if err != nil {
		return false, err
	}
	return resp.WorkflowExecutionInfo.CloseStatus != nil, nil
}



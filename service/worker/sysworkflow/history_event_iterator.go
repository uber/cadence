package sysworkflow

import (
	"context"
	"errors"
	"github.com/uber/cadence/client/public"
	"go.uber.org/cadence/.gen/go/shared"
)

type (
	// HistoryEventIterator gets history events using public.Client
	HistoryEventIterator interface {
		Next() (*shared.HistoryEvent, error)
		HasNext() bool
	}

	historyEventIterator struct {
		domain *string
		execution *shared.WorkflowExecution
		publicClient public.Client
		buffer []*shared.HistoryEvent
		bufferIndex int
		nextPageToken []byte
		hasNextPage bool
	}
)

// NewHistoryEventIterator returns a new HistoryEventIterator
func NewHistoryEventIterator(
	domain *string,
	execution *shared.WorkflowExecution,
	publicClient public.Client,
) HistoryEventIterator {
	return &historyEventIterator{
		domain: domain,
		execution: execution,
		publicClient: publicClient,
		hasNextPage: true,
	}
}

// Next returns next history event. Returns error on failure to read page from public.Client or if iterator is finished.
func (i *historyEventIterator) Next() (*shared.HistoryEvent, error) {
	if !i.HasNext() {
		return nil, errors.New("iterator is empty")
	}

	// if buffer has not yet been loaded for the first time or if current buffer is used up, then get the next page
	if len(i.buffer) == 0 || i.bufferIndex >= len(i.buffer) {
		request := &shared.GetWorkflowExecutionHistoryRequest{
			Domain: i.domain,
			Execution: i.execution,
			NextPageToken: i.nextPageToken,
		}
		resp, err := i.publicClient.GetWorkflowExecutionHistory(context.Background(), request)
		if err != nil {
			// TODO: think through how this works with retryable errors
			return nil, err
		}
		i.buffer = resp.History.Events
		i.bufferIndex = 0
		i.nextPageToken = resp.NextPageToken
		if resp.NextPageToken == nil {
			i.hasNextPage = false
		}
	}
	event := i.buffer[i.bufferIndex]
	i.bufferIndex++
	return event, nil
}

// HasNext returns true if there are more items to iterate through
func (i *historyEventIterator) HasNext() bool {
	return i.hasNextPage || i.bufferIndex < len(i.buffer)
}

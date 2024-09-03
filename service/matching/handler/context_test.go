package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cadence_errors "github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

func TestHandleErrNil(t *testing.T) {
	reqCtx := &handlerContext{
		scope:  metrics.NoopScope(metrics.Matching),
		logger: log.NewNoop(),
	}

	err := reqCtx.handleErr(nil)
	assert.NoError(t, err)
}

func TestHandleErrKnowErrors(t *testing.T) {
	testCases := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name: "InternalServiceError",
			err:  &types.InternalServiceError{},
		},
		{
			name: "BadRequestError",
			err:  &types.BadRequestError{},
		},
		{
			name: "EntityNotExistsError",
			err:  &types.EntityNotExistsError{},
		},
		{
			name: "WorkflowExecutionAlreadyStartedError",
			err:  &types.WorkflowExecutionAlreadyStartedError{},
		},
		{
			name: "DomainAlreadyExistsError",
			err:  &types.DomainAlreadyExistsError{},
		},
		{
			name: "QueryFailedError",
			err:  &types.QueryFailedError{},
		},
		{
			name: "LimitExceededError",
			err:  &types.LimitExceededError{},
		},
		{
			name: "ServiceBusyError",
			err:  &types.ServiceBusyError{},
		},
		{
			name: "DomainNotActiveError",
			err:  &types.DomainNotActiveError{},
		},
		{
			name: "RemoteSyncMatchedError",
			err:  &types.RemoteSyncMatchedError{},
		},
		{
			name: "StickyWorkerUnavailableError",
			err:  &types.StickyWorkerUnavailableError{},
		},
		{
			name: "TaskListNotOwnedByHostError",
			err:  &cadence_errors.TaskListNotOwnedByHostError{},
		},
	}

	for _, tc := range testCases {
		reqCtx := &handlerContext{
			scope:  metrics.NoopScope(metrics.Matching),
			logger: log.NewNoop(),
		}

		t.Run(tc.name, func(t *testing.T) {
			handledErr := reqCtx.handleErr(tc.err)
			require.Equal(t, tc.err, handledErr)

			wrappedErr := fmt.Errorf("wrapped: %w", tc.err)
			wrappedHandledErr := reqCtx.handleErr(wrappedErr)
			require.Equal(t, wrappedErr, wrappedHandledErr)
		})
	}
}

func TestHandleErrUncategorizedError(t *testing.T) {
	reqCtx := &handlerContext{
		scope:  metrics.NoopScope(metrics.Matching),
		logger: log.NewNoop(),
	}

	err := reqCtx.handleErr(assert.AnError)
	assert.ErrorAs(t, err, new(*types.InternalServiceError))
}

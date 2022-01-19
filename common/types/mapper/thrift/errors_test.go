// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package thrift

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

func TestFromErrorMapping(t *testing.T) {

	tests := map[string]struct {
		in          error
		expectedOut error
	}{
		"generic error": {
			in:          assert.AnError,
			expectedOut: assert.AnError,
		},
		"&types.AccessDeniedError": {
			in:          &types.AccessDeniedError{},
			expectedOut: &shared.AccessDeniedError{},
		},
		"&types.BadRequestError": {
			in:          &types.BadRequestError{},
			expectedOut: &shared.BadRequestError{},
		},
		"&types.CancellationAlreadyRequestedError": {
			in:          &types.CancellationAlreadyRequestedError{},
			expectedOut: &shared.CancellationAlreadyRequestedError{},
		},
		"&types.ClientVersionNotSupportedError": {
			in:          &types.ClientVersionNotSupportedError{},
			expectedOut: &shared.ClientVersionNotSupportedError{},
		},
		"&types.FeatureNotEnabledError": {
			in:          &types.FeatureNotEnabledError{},
			expectedOut: &shared.FeatureNotEnabledError{},
		},
		"&types.CurrentBranchChangedError": {
			in:          &types.CurrentBranchChangedError{},
			expectedOut: &shared.CurrentBranchChangedError{},
		},
		"&types.DomainAlreadyExistsError": {
			in:          &types.DomainAlreadyExistsError{},
			expectedOut: &shared.DomainAlreadyExistsError{},
		},
		"&types.DomainNotActiveError": {
			in:          &types.DomainNotActiveError{},
			expectedOut: &shared.DomainNotActiveError{},
		},
		"&types.EntityNotExistsError": {
			in:          &types.EntityNotExistsError{},
			expectedOut: &shared.EntityNotExistsError{},
		},
		"&types.WorkflowExecutionAlreadyCompletedError": {
			in:          &types.WorkflowExecutionAlreadyCompletedError{},
			expectedOut: &shared.WorkflowExecutionAlreadyCompletedError{},
		},
		"&types.InternalDataInconsistencyError": {
			in:          &types.InternalDataInconsistencyError{},
			expectedOut: &shared.InternalDataInconsistencyError{},
		},
		"&types.InternalServiceError": {
			in:          &types.InternalServiceError{},
			expectedOut: &shared.InternalServiceError{},
		},
		"&types.LimitExceededError": {
			in:          &types.LimitExceededError{},
			expectedOut: &shared.LimitExceededError{},
		},
		"&types.QueryFailedError": {
			in:          &types.QueryFailedError{},
			expectedOut: &shared.QueryFailedError{},
		},
		"&types.RemoteSyncMatchedError": {
			in:          &types.RemoteSyncMatchedError{},
			expectedOut: &shared.RemoteSyncMatchedError{},
		},
		"&types.RetryTaskV2Error": {
			in:          &types.RetryTaskV2Error{},
			expectedOut: &shared.RetryTaskV2Error{},
		},
		"&types.ServiceBusyError": {
			in:          &types.ServiceBusyError{},
			expectedOut: &shared.ServiceBusyError{},
		},
		"&types.WorkflowExecutionAlreadyStartedError": {
			in:          &types.WorkflowExecutionAlreadyStartedError{},
			expectedOut: &shared.WorkflowExecutionAlreadyStartedError{},
		},
		"&types.ShardOwnershipLostError": {
			in:          &types.ShardOwnershipLostError{},
			expectedOut: &history.ShardOwnershipLostError{},
		},
		"&types.EventAlreadyStartedError": {
			in:          &types.EventAlreadyStartedError{},
			expectedOut: &history.EventAlreadyStartedError{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, errors.As(FromError(td.in), &td.expectedOut))
		})
	}
}

func TestToErrorMapping(t *testing.T) {

	tests := map[string]struct {
		in          error
		expectedOut error
	}{
		"generic error": {
			in:          assert.AnError,
			expectedOut: assert.AnError,
		},
		"&types.AccessDeniedError": {
			in:          &shared.AccessDeniedError{},
			expectedOut: &types.AccessDeniedError{},
		},
		"&types.BadRequestError": {
			in:          &shared.BadRequestError{},
			expectedOut: &types.BadRequestError{},
		},
		"&types.CancellationAlreadyRequestedError": {
			in:          &shared.CancellationAlreadyRequestedError{},
			expectedOut: &types.CancellationAlreadyRequestedError{},
		},
		"&types.ClientVersionNotSupportedError": {
			in:          &shared.ClientVersionNotSupportedError{},
			expectedOut: &types.ClientVersionNotSupportedError{},
		},
		"&types.FeatureNotEnabledError": {
			in:          &shared.FeatureNotEnabledError{},
			expectedOut: &types.FeatureNotEnabledError{},
		},
		"&types.CurrentBranchChangedError": {
			in:          &shared.CurrentBranchChangedError{},
			expectedOut: &types.CurrentBranchChangedError{},
		},
		"&types.DomainAlreadyExistsError": {
			in:          &shared.DomainAlreadyExistsError{},
			expectedOut: &types.DomainAlreadyExistsError{},
		},
		"&types.DomainNotActiveError": {
			in:          &shared.DomainNotActiveError{},
			expectedOut: &types.DomainNotActiveError{},
		},
		"&types.EntityNotExistsError": {
			in:          &shared.EntityNotExistsError{},
			expectedOut: &types.EntityNotExistsError{},
		},
		"&types.WorkflowExecutionAlreadyCompletedError": {
			in:          &shared.WorkflowExecutionAlreadyCompletedError{},
			expectedOut: &types.WorkflowExecutionAlreadyCompletedError{},
		},
		"&types.InternalDataInconsistencyError": {
			in:          &shared.InternalDataInconsistencyError{},
			expectedOut: &types.InternalDataInconsistencyError{},
		},
		"&types.InternalServiceError": {
			in:          &shared.InternalServiceError{},
			expectedOut: &types.InternalServiceError{},
		},
		"&types.LimitExceededError": {
			in:          &shared.LimitExceededError{},
			expectedOut: &types.LimitExceededError{},
		},
		"&types.QueryFailedError": {
			in:          &shared.QueryFailedError{},
			expectedOut: &types.QueryFailedError{},
		},
		"&types.RemoteSyncMatchedError": {
			in:          &shared.RemoteSyncMatchedError{},
			expectedOut: &types.RemoteSyncMatchedError{},
		},
		"&types.RetryTaskV2Error": {
			in:          &shared.RetryTaskV2Error{},
			expectedOut: &types.RetryTaskV2Error{},
		},
		"&types.ServiceBusyError": {
			in:          &shared.ServiceBusyError{},
			expectedOut: &types.ServiceBusyError{},
		},
		"&types.WorkflowExecutionAlreadyStartedError": {
			in:          &shared.WorkflowExecutionAlreadyStartedError{},
			expectedOut: &types.WorkflowExecutionAlreadyStartedError{},
		},
		"&types.ShardOwnershipLostError": {
			in:          &history.ShardOwnershipLostError{},
			expectedOut: &history.ShardOwnershipLostError{},
		},
		"&types.EventAlreadyStartedError": {
			in:          &history.EventAlreadyStartedError{},
			expectedOut: &history.EventAlreadyStartedError{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, errors.As(ToError(td.in), &td.expectedOut))
		})
	}
}

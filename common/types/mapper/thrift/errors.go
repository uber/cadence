// Copyright (c) 2020 Uber Technologies Inc.
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
	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

// FromError convert error to Thrift type if it comes as its internal equivalent
func FromError(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *types.AccessDeniedError:
		return FromAccessDeniedError(e)
	case *types.BadRequestError:
		return FromBadRequestError(e)
	case *types.CancellationAlreadyRequestedError:
		return FromCancellationAlreadyRequestedError(e)
	case *types.ClientVersionNotSupportedError:
		return FromClientVersionNotSupportedError(e)
	case *types.CurrentBranchChangedError:
		return FromCurrentBranchChangedError(e)
	case *types.DomainAlreadyExistsError:
		return FromDomainAlreadyExistsError(e)
	case *types.DomainNotActiveError:
		return FromDomainNotActiveError(e)
	case *types.EntityNotExistsError:
		return FromEntityNotExistsError(e)
	case *types.WorkflowExecutionAlreadyCompletedError:
		return FromWorkflowExecutionAlreadyCompletedError(e)
	case *types.InternalDataInconsistencyError:
		return FromInternalDataInconsistencyError(e)
	case *types.InternalServiceError:
		return FromInternalServiceError(e)
	case *types.LimitExceededError:
		return FromLimitExceededError(e)
	case *types.QueryFailedError:
		return FromQueryFailedError(e)
	case *types.RemoteSyncMatchedError:
		return FromRemoteSyncMatchedError(e)
	case *types.RetryTaskV2Error:
		return FromRetryTaskV2Error(e)
	case *types.ServiceBusyError:
		return FromServiceBusyError(e)
	case *types.WorkflowExecutionAlreadyStartedError:
		return FromWorkflowExecutionAlreadyStartedError(e)
	case *types.ShardOwnershipLostError:
		return FromShardOwnershipLostError(e)
	case *types.EventAlreadyStartedError:
		return FromEventAlreadyStartedError(e)
	default:
		return err
	}
}

// ToError convert error to internal type if it comes as its thrift equivalent
func ToError(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *shared.AccessDeniedError:
		return ToAccessDeniedError(e)
	case *shared.BadRequestError:
		return ToBadRequestError(e)
	case *shared.CancellationAlreadyRequestedError:
		return ToCancellationAlreadyRequestedError(e)
	case *shared.ClientVersionNotSupportedError:
		return ToClientVersionNotSupportedError(e)
	case *shared.CurrentBranchChangedError:
		return ToCurrentBranchChangedError(e)
	case *shared.DomainAlreadyExistsError:
		return ToDomainAlreadyExistsError(e)
	case *shared.DomainNotActiveError:
		return ToDomainNotActiveError(e)
	case *shared.EntityNotExistsError:
		return ToEntityNotExistsError(e)
	case *shared.WorkflowExecutionAlreadyCompletedError:
		return ToWorkflowExecutionAlreadyCompletedError(e)
	case *shared.InternalDataInconsistencyError:
		return ToInternalDataInconsistencyError(e)
	case *shared.InternalServiceError:
		return ToInternalServiceError(e)
	case *shared.LimitExceededError:
		return ToLimitExceededError(e)
	case *shared.QueryFailedError:
		return ToQueryFailedError(e)
	case *shared.RemoteSyncMatchedError:
		return ToRemoteSyncMatchedError(e)
	case *shared.RetryTaskV2Error:
		return ToRetryTaskV2Error(e)
	case *shared.ServiceBusyError:
		return ToServiceBusyError(e)
	case *shared.WorkflowExecutionAlreadyStartedError:
		return ToWorkflowExecutionAlreadyStartedError(e)
	case *history.ShardOwnershipLostError:
		return ToShardOwnershipLostError(e)
	case *history.EventAlreadyStartedError:
		return ToEventAlreadyStartedError(e)
	default:
		return err
	}
}

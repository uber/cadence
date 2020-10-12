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

import "github.com/uber/cadence/common/types"

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
	default:
		return err
	}
}

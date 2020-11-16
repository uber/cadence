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

package types

import (
	"fmt"
	"strings"
)

func (err AccessDeniedError) Error() string {
	return fmt.Sprintf("AccessDeniedError{Message: %v}", err.Message)
}

func (err BadRequestError) Error() string {
	return fmt.Sprintf("BadRequestError{Message: %v}", err.Message)
}

func (err CancellationAlreadyRequestedError) Error() string {
	return fmt.Sprintf("CancellationAlreadyRequestedError{Message: %v}", err.Message)
}

func (err ClientVersionNotSupportedError) Error() string {
	return fmt.Sprintf("ClientVersionNotSupportedError{FeatureVersion: %v, ClientImpl: %v, SupportedVersions: %v}",
		err.FeatureVersion,
		err.ClientImpl,
		err.SupportedVersions)
}

func (err CurrentBranchChangedError) Error() string {
	return fmt.Sprintf("CurrentBranchChangedError{Message: %v, CurrentBranchToken: %v}",
		err.Message,
		err.CurrentBranchToken)
}

func (err DomainAlreadyExistsError) Error() string {
	return fmt.Sprintf("DomainAlreadyExistsError{Message: %v}", err.Message)
}

func (err DomainNotActiveError) Error() string {
	return fmt.Sprintf("DomainNotActiveError{Message: %v, DomainName: %v, CurrentCluster: %v, ActiveCluster: %v}",
		err.Message,
		err.DomainName,
		err.CurrentCluster,
		err.ActiveCluster,
	)
}

func (err EntityNotExistsError) Error() string {
	fields := []string{}
	fields = append(fields, fmt.Sprintf("Message: %v", err.Message))
	if err.CurrentCluster != nil {
		fields = append(fields, fmt.Sprintf("CurrentCluster: %v", err.CurrentCluster))
	}
	if err.ActiveCluster != nil {
		fields = append(fields, fmt.Sprintf("ActiveCluster: %v", err.ActiveCluster))
	}
	return fmt.Sprintf("EntityNotExistsError{%s}", strings.Join(fields, ", "))
}

func (err InternalDataInconsistencyError) Error() string {
	return fmt.Sprintf("InternalDataInconsistencyError{Message: %v}", err.Message)
}

func (err InternalServiceError) Error() string {
	return fmt.Sprintf("InternalServiceError{Message: %v}", err.Message)
}

func (err LimitExceededError) Error() string {
	return fmt.Sprintf("LimitExceededError{Message: %v}", err.Message)
}

func (err QueryFailedError) Error() string {
	return fmt.Sprintf("QueryFailedError{Message: %v}", err.Message)
}

func (err RemoteSyncMatchedError) Error() string {
	return fmt.Sprintf("RemoteSyncMatchedError{Message: %v}", err.Message)
}

func (err RetryTaskV2Error) Error() string {
	fields := []string{}
	fields = append(fields, fmt.Sprintf("Message: %v", err.Message))
	if err.DomainID != nil {
		fields = append(fields, fmt.Sprintf("DomainID: %v", err.DomainID))
	}
	if err.WorkflowID != nil {
		fields = append(fields, fmt.Sprintf("WorkflowID: %v", err.WorkflowID))
	}
	if err.RunID != nil {
		fields = append(fields, fmt.Sprintf("RunID: %v", err.RunID))
	}
	if err.StartEventID != nil {
		fields = append(fields, fmt.Sprintf("StartEventID: %v", err.StartEventID))
	}
	if err.StartEventVersion != nil {
		fields = append(fields, fmt.Sprintf("StartEventVersion: %v", err.StartEventVersion))
	}
	if err.EndEventID != nil {
		fields = append(fields, fmt.Sprintf("EndEventID: %v", err.EndEventID))
	}
	if err.EndEventVersion != nil {
		fields = append(fields, fmt.Sprintf("EndEventVersion: %v", err.EndEventVersion))
	}
	return fmt.Sprintf("RetryTaskV2Error{%s}", strings.Join(fields, ", "))
}

func (err ServiceBusyError) Error() string {
	return fmt.Sprintf("ServiceBusyError{Message: %v}", err.Message)
}

func (err WorkflowExecutionAlreadyStartedError) Error() string {
	fields := []string{}
	if err.Message != nil {
		fields = append(fields, fmt.Sprintf("Message: %v", err.Message))
	}
	if err.StartRequestID != nil {
		fields = append(fields, fmt.Sprintf("StartRequestID: %v", err.StartRequestID))
	}
	if err.RunID != nil {
		fields = append(fields, fmt.Sprintf("RunID: %v", err.RunID))
	}
	return fmt.Sprintf("WorkflowExecutionAlreadyStartedError{%s}", strings.Join(fields, ", "))
}

func (err ShardOwnershipLostError) Error() string {
	fields := []string{}
	if err.Message != nil {
		fields = append(fields, fmt.Sprintf("Message: %v", err.Message))
	}
	if err.Owner != nil {
		fields = append(fields, fmt.Sprintf("StartRequestID: %v", err.Owner))
	}
	return fmt.Sprintf("ShardOwnershipLostError{%s}", strings.Join(fields, ", "))
}

func (err EventAlreadyStartedError) Error() string {
	return fmt.Sprintf("EventAlreadyStartedError{Message: %v}", err.Message)
}

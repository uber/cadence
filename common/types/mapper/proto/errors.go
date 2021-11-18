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

package proto

import (
	"errors"

	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	sharedv1 "github.com/uber/cadence/.gen/proto/shared/v1"
	"github.com/uber/cadence/common/types"
)

func FromError(err error) error {
	if err == nil {
		return protobuf.NewError(yarpcerrors.CodeOK, "")
	}

	switch e := err.(type) {
	case *types.AccessDeniedError:
		return protobuf.NewError(yarpcerrors.CodePermissionDenied, e.Message)
	case *types.InternalServiceError:
		return protobuf.NewError(yarpcerrors.CodeInternal, e.Message)
	case *types.EntityNotExistsError:
		return protobuf.NewError(yarpcerrors.CodeNotFound, e.Message, protobuf.WithErrorDetails(&apiv1.EntityNotExistsError{
			CurrentCluster: e.CurrentCluster,
			ActiveCluster:  e.ActiveCluster,
		}))
	case *types.WorkflowExecutionAlreadyCompletedError:
		return protobuf.NewError(yarpcerrors.CodeNotFound, e.Message, protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyCompletedError{}))
	case *types.BadRequestError:
		return protobuf.NewError(yarpcerrors.CodeInvalidArgument, e.Message)
	case *types.QueryFailedError:
		return protobuf.NewError(yarpcerrors.CodeInvalidArgument, e.Message, protobuf.WithErrorDetails(&apiv1.QueryFailedError{}))
	case *types.ShardOwnershipLostError:
		return protobuf.NewError(yarpcerrors.CodeAborted, e.Message, protobuf.WithErrorDetails(&sharedv1.ShardOwnershipLostError{
			Owner: e.Owner,
		}))
	case *types.CurrentBranchChangedError:
		return protobuf.NewError(yarpcerrors.CodeAborted, e.Message, protobuf.WithErrorDetails(&sharedv1.CurrentBranchChangedError{
			CurrentBranchToken: e.GetCurrentBranchToken(),
		}))
	case *types.RetryTaskV2Error:
		return protobuf.NewError(yarpcerrors.CodeAborted, e.Message, protobuf.WithErrorDetails(&sharedv1.RetryTaskV2Error{
			DomainId:          e.DomainID,
			WorkflowExecution: FromWorkflowRunPair(e.WorkflowID, e.RunID),
			StartEvent:        FromEventIDVersionPair(e.StartEventID, e.StartEventVersion),
			EndEvent:          FromEventIDVersionPair(e.EndEventID, e.EndEventVersion),
		}))
	case *types.CancellationAlreadyRequestedError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&apiv1.CancellationAlreadyRequestedError{}))
	case *types.DomainAlreadyExistsError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&apiv1.DomainAlreadyExistsError{}))
	case *types.EventAlreadyStartedError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&sharedv1.EventAlreadyStartedError{}))
	case *types.WorkflowExecutionAlreadyStartedError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyStartedError{
			StartRequestId: e.StartRequestID,
			RunId:          e.RunID,
		}))
	case *types.ClientVersionNotSupportedError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Client version not supported", protobuf.WithErrorDetails(&apiv1.ClientVersionNotSupportedError{
			FeatureVersion:    e.FeatureVersion,
			ClientImpl:        e.ClientImpl,
			SupportedVersions: e.SupportedVersions,
		}))
	case *types.FeatureNotEnabledError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Feature flag not enabled", protobuf.WithErrorDetails(&apiv1.FeatureNotEnabledError{
			FeatureFlag: e.FeatureFlag,
		}))
	case *types.DomainNotActiveError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, e.Message, protobuf.WithErrorDetails(&apiv1.DomainNotActiveError{
			Domain:         e.DomainName,
			CurrentCluster: e.CurrentCluster,
			ActiveCluster:  e.ActiveCluster,
		}))
	case *types.InternalDataInconsistencyError:
		return protobuf.NewError(yarpcerrors.CodeDataLoss, e.Message, protobuf.WithErrorDetails(&sharedv1.InternalDataInconsistencyError{}))
	case *types.LimitExceededError:
		return protobuf.NewError(yarpcerrors.CodeResourceExhausted, e.Message, protobuf.WithErrorDetails(&apiv1.LimitExceededError{}))
	case *types.ServiceBusyError:
		return protobuf.NewError(yarpcerrors.CodeResourceExhausted, e.Message, protobuf.WithErrorDetails(&apiv1.ServiceBusyError{}))
	case *types.RemoteSyncMatchedError:
		return protobuf.NewError(yarpcerrors.CodeUnavailable, e.Message, protobuf.WithErrorDetails(&sharedv1.RemoteSyncMatchedError{}))
	}

	return protobuf.NewError(yarpcerrors.CodeUnknown, err.Error())
}

func ToError(err error) error {
	status := yarpcerrors.FromError(err)
	if status == nil || status.Code() == yarpcerrors.CodeOK {
		return nil
	}

	switch status.Code() {
	case yarpcerrors.CodePermissionDenied:
		return &types.AccessDeniedError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeInternal:
		return &types.InternalServiceError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeNotFound:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.EntityNotExistsError:
			return &types.EntityNotExistsError{
				Message:        status.Message(),
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		case *apiv1.WorkflowExecutionAlreadyCompletedError:
			return &types.WorkflowExecutionAlreadyCompletedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeInvalidArgument:
		switch getErrorDetails(err).(type) {
		case nil:
			return &types.BadRequestError{
				Message: status.Message(),
			}
		case *apiv1.QueryFailedError:
			return &types.QueryFailedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeAborted:
		switch details := getErrorDetails(err).(type) {
		case *sharedv1.ShardOwnershipLostError:
			return &types.ShardOwnershipLostError{
				Message: status.Message(),
				Owner:   details.Owner,
			}
		case *sharedv1.CurrentBranchChangedError:
			return &types.CurrentBranchChangedError{
				Message:            status.Message(),
				CurrentBranchToken: details.CurrentBranchToken,
			}
		case *sharedv1.RetryTaskV2Error:
			return &types.RetryTaskV2Error{
				Message:           status.Message(),
				DomainID:          details.DomainId,
				WorkflowID:        ToWorkflowID(details.WorkflowExecution),
				RunID:             ToRunID(details.WorkflowExecution),
				StartEventID:      ToEventID(details.StartEvent),
				StartEventVersion: ToEventVersion(details.StartEvent),
				EndEventID:        ToEventID(details.EndEvent),
				EndEventVersion:   ToEventVersion(details.EndEvent),
			}
		}
	case yarpcerrors.CodeAlreadyExists:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.CancellationAlreadyRequestedError:
			return &types.CancellationAlreadyRequestedError{
				Message: status.Message(),
			}
		case *apiv1.DomainAlreadyExistsError:
			return &types.DomainAlreadyExistsError{
				Message: status.Message(),
			}
		case *sharedv1.EventAlreadyStartedError:
			return &types.EventAlreadyStartedError{
				Message: status.Message(),
			}
		case *apiv1.WorkflowExecutionAlreadyStartedError:
			return &types.WorkflowExecutionAlreadyStartedError{
				Message:        status.Message(),
				StartRequestID: details.StartRequestId,
				RunID:          details.RunId,
			}
		}
	case yarpcerrors.CodeDataLoss:
		return &types.InternalDataInconsistencyError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeFailedPrecondition:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.ClientVersionNotSupportedError:
			return &types.ClientVersionNotSupportedError{
				FeatureVersion:    details.FeatureVersion,
				ClientImpl:        details.ClientImpl,
				SupportedVersions: details.SupportedVersions,
			}
		case *apiv1.FeatureNotEnabledError:
			return &types.FeatureNotEnabledError{
				FeatureFlag: details.FeatureFlag,
			}
		case *apiv1.DomainNotActiveError:
			return &types.DomainNotActiveError{
				Message:        status.Message(),
				DomainName:     details.Domain,
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		}
	case yarpcerrors.CodeResourceExhausted:
		switch getErrorDetails(err).(type) {
		case *apiv1.LimitExceededError:
			return &types.LimitExceededError{
				Message: status.Message(),
			}
		case *apiv1.ServiceBusyError:
			return &types.ServiceBusyError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeUnavailable:
		switch getErrorDetails(err).(type) {
		case *sharedv1.RemoteSyncMatchedError:
			return &types.RemoteSyncMatchedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeUnknown:
		return errors.New(status.Message())
	}

	// If error does not match anything, return raw yarpc status error
	// There are some code that casts error to yarpc status to check for deadline exceeded status
	return status
}

func getErrorDetails(err error) interface{} {
	details := protobuf.GetErrorDetails(err)
	if len(details) > 0 {
		return details[0]
	}
	return nil
}

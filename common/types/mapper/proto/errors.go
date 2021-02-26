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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/uber/cadence/.gen/proto/api/v1"
	sharedv1 "github.com/uber/cadence/.gen/proto/shared/v1"
	"github.com/uber/cadence/common/types"
)

func FromError(err error) *status.Status {
	st, conversionErr := errorToStatus(err)
	if conversionErr != nil {
		return status.Newf(codes.Internal, "failed to convert error to proto status: %s", err.Error())
	}
	return st
}

func errorToStatus(err error) (*status.Status, error) {
	if err == nil {
		return status.New(codes.OK, ""), nil
	}

	switch e := err.(type) {
	case types.AccessDeniedError:
		return status.New(codes.PermissionDenied, e.Message), nil
	case types.InternalServiceError:
		return status.New(codes.Internal, e.Message), nil
	case types.EntityNotExistsError:
		return status.New(codes.NotFound, e.Message).WithDetails(&apiv1.EntityNotExistsError{
			CurrentCluster: e.CurrentCluster,
			ActiveCluster:  e.ActiveCluster,
		})
	case types.BadRequestError:
		return status.New(codes.InvalidArgument, e.Message), nil
	case types.QueryFailedError:
		return status.New(codes.InvalidArgument, e.Message).WithDetails(&apiv1.QueryFailedError{})
	case types.ShardOwnershipLostError:
		return status.New(codes.Aborted, e.Message).WithDetails(&sharedv1.ShardOwnershipLostError{
			Owner: e.Owner,
		})
	case types.CurrentBranchChangedError:
		return status.New(codes.Aborted, e.Message).WithDetails(&sharedv1.CurrentBranchChangedError{
			CurrentBranchToken: e.GetCurrentBranchToken(),
		})
	case types.RetryTaskV2Error:
		return status.New(codes.Aborted, e.Message).WithDetails(&sharedv1.RetryTaskV2Error{
			DomainId:          e.DomainID,
			WorkflowExecution: FromWorkflowRunPair(e.WorkflowID, e.RunID),
			StartEvent:        FromEventIDVersionPair(e.StartEventID, e.StartEventVersion),
			EndEvent:          FromEventIDVersionPair(e.EndEventID, e.EndEventVersion),
		})
	case types.CancellationAlreadyRequestedError:
		return status.New(codes.AlreadyExists, e.Message).WithDetails(&apiv1.CancellationAlreadyRequestedError{})
	case types.DomainAlreadyExistsError:
		return status.New(codes.AlreadyExists, e.Message).WithDetails(&apiv1.DomainAlreadyExistsError{})
	case types.EventAlreadyStartedError:
		return status.New(codes.AlreadyExists, e.Message).WithDetails(&sharedv1.EventAlreadyStartedError{})
	case types.WorkflowExecutionAlreadyStartedError:
		return status.New(codes.AlreadyExists, e.Message).WithDetails(&apiv1.WorkflowExecutionAlreadyStartedError{
			StartRequestId: e.StartRequestID,
			RunId:          e.RunID,
		})
	case types.ClientVersionNotSupportedError:
		return status.New(codes.FailedPrecondition, "Client version not supported").WithDetails(&apiv1.ClientVersionNotSupportedError{
			FeatureVersion:    e.FeatureVersion,
			ClientImpl:        e.ClientImpl,
			SupportedVersions: e.SupportedVersions,
		})
	case types.DomainNotActiveError:
		return status.New(codes.FailedPrecondition, e.Message).WithDetails(&apiv1.DomainNotActiveError{
			Domain:         e.DomainName,
			CurrentCluster: e.CurrentCluster,
			ActiveCluster:  e.ActiveCluster,
		})
	case types.InternalDataInconsistencyError:
		return status.New(codes.DataLoss, e.Message).WithDetails(&sharedv1.InternalDataInconsistencyError{})
	case types.LimitExceededError:
		return status.New(codes.ResourceExhausted, e.Message).WithDetails(&apiv1.LimitExceededError{})
	case types.ServiceBusyError:
		return status.New(codes.ResourceExhausted, e.Message).WithDetails(&apiv1.ServiceBusyError{})
	case types.RemoteSyncMatchedError:
		return status.New(codes.Unavailable, e.Message).WithDetails(&sharedv1.RemoteSyncMatchedError{})
	}

	return status.New(codes.Unknown, err.Error()), nil
}

func ToError(status *status.Status) error {
	if status == nil || status.Code() == codes.OK {
		return nil
	}

	switch status.Code() {
	case codes.PermissionDenied:
		return types.AccessDeniedError{
			Message: status.Message(),
		}
	case codes.Internal:
		return types.InternalServiceError{
			Message: status.Message(),
		}
	case codes.NotFound:
		switch details := getErrorDetails(status).(type) {
		case *apiv1.EntityNotExistsError:
			return types.EntityNotExistsError{
				Message:        status.Message(),
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		}
	case codes.InvalidArgument:
		switch getErrorDetails(status).(type) {
		case nil:
			return types.BadRequestError{
				Message: status.Message(),
			}
		case *apiv1.QueryFailedError:
			return types.QueryFailedError{
				Message: status.Message(),
			}
		}
	case codes.Aborted:
		switch details := getErrorDetails(status).(type) {
		case *sharedv1.ShardOwnershipLostError:
			return types.ShardOwnershipLostError{
				Message: status.Message(),
				Owner:   details.Owner,
			}
		case *sharedv1.CurrentBranchChangedError:
			return types.CurrentBranchChangedError{
				Message:            status.Message(),
				CurrentBranchToken: details.CurrentBranchToken,
			}
		case *sharedv1.RetryTaskV2Error:
			return types.RetryTaskV2Error{
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
	case codes.AlreadyExists:
		switch details := getErrorDetails(status).(type) {
		case *apiv1.CancellationAlreadyRequestedError:
			return types.CancellationAlreadyRequestedError{
				Message: status.Message(),
			}
		case *apiv1.DomainAlreadyExistsError:
			return types.DomainAlreadyExistsError{
				Message: status.Message(),
			}
		case *sharedv1.EventAlreadyStartedError:
			return types.EventAlreadyStartedError{
				Message: status.Message(),
			}
		case *apiv1.WorkflowExecutionAlreadyStartedError:
			return types.WorkflowExecutionAlreadyStartedError{
				Message:        status.Message(),
				StartRequestID: details.StartRequestId,
				RunID:          details.RunId,
			}
		}
	case codes.DataLoss:
		return types.InternalDataInconsistencyError{
			Message: status.Message(),
		}
	case codes.FailedPrecondition:
		switch details := getErrorDetails(status).(type) {
		case *apiv1.ClientVersionNotSupportedError:
			return types.ClientVersionNotSupportedError{
				FeatureVersion:    details.FeatureVersion,
				ClientImpl:        details.ClientImpl,
				SupportedVersions: details.SupportedVersions,
			}
		case *apiv1.DomainNotActiveError:
			return types.DomainNotActiveError{
				Message:        status.Message(),
				DomainName:     details.Domain,
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		}
	case codes.ResourceExhausted:
		switch getErrorDetails(status).(type) {
		case *apiv1.LimitExceededError:
			return types.LimitExceededError{
				Message: status.Message(),
			}
		case *apiv1.ServiceBusyError:
			return types.ServiceBusyError{
				Message: status.Message(),
			}
		}
	case codes.Unavailable:
		switch getErrorDetails(status).(type) {
		case *sharedv1.RemoteSyncMatchedError:
			return types.RemoteSyncMatchedError{
				Message: status.Message(),
			}
		}
	}

	return errors.New(status.Message())
}

func getErrorDetails(status *status.Status) interface{} {
	details := status.Details()
	if len(details) > 0 {
		return details[0]
	}
	return nil
}

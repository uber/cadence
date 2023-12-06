package errors

import (
	"errors"
	"github.com/uber/cadence/common/types"
)

// IsShardOwnershipLostErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsShardOwnershipLostErrorType(err error) bool {
	e1 := &types.ShardOwnershipLostError{}
	return errors.As(err, &e1)
}

// IsEventAlreadyStartedErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsEventAlreadyStartedErrorType(err error) bool {
	e1 := &types.EventAlreadyStartedError{}
	return errors.As(err, &e1)
}

// IsBadRequestErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsBadRequestErrorType(err error) bool {
	e1 := &types.BadRequestError{}
	return errors.As(err, &e1)
}

// IsDomainNotActiveErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsDomainNotActiveErrorType(err error) bool {
	e1 := &types.DomainNotActiveError{}
	return errors.As(err, &e1)
}

// IsWorkflowExecutionAlreadyStartedErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsWorkflowExecutionAlreadyStartedErrorType(err error) bool {
	e1 := &types.WorkflowExecutionAlreadyStartedError{}
	return errors.As(err, &e1)
}

// IsEntityNotExistsErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsEntityNotExistsErrorType(err error) bool {
	e1 := &types.EntityNotExistsError{}
	return errors.As(err, &e1)
}

// IsWorkflowExecutionAlreadyCompletedErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsWorkflowExecutionAlreadyCompletedErrorType(err error) bool {
	e1 := &types.WorkflowExecutionAlreadyCompletedError{}
	return errors.As(err, &e1)
}

// IsCancellationAlreadyRequestedErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsCancellationAlreadyRequestedErrorType(err error) bool {
	e1 := &types.CancellationAlreadyRequestedError{}
	return errors.As(err, &e1)
}

// IsLimitExceededErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsLimitExceededErrorType(err error) bool {
	e1 := &types.LimitExceededError{}
	return errors.As(err, &e1)
}

// IsRetryTaskV2ErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsRetryTaskV2ErrorType(err error) bool {
	e1 := &types.RetryTaskV2Error{}
	return errors.As(err, &e1)
}

// IsServiceBusyErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsServiceBusyErrorType(err error) bool {
	e1 := &types.ServiceBusyError{}
	return errors.As(err, &e1)
}

// IsInternalServiceErrorType is a convenience wrapper for errors.As, which takes in an error and returns true if they are of the same type.
func IsInternalServiceErrorType(err error) bool {
	e1 := &types.InternalServiceError{}
	return errors.As(err, &e1)
}

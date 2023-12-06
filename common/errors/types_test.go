package errors

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/types"
	"testing"
)

func TestIsShardOwnershipLostErrorType(t *testing.T) {
	assert.False(t, IsShardOwnershipLostErrorType(errors.New("a random error")))
	assert.True(t, IsShardOwnershipLostErrorType(&types.ShardOwnershipLostError{Message: "test"}))
}

func TestBadRequestErrorType(t *testing.T) {
	assert.False(t, IsBadRequestErrorType(errors.New("a random error")))
	assert.True(t, IsBadRequestErrorType(&types.BadRequestError{Message: "test"}))
}

func TestDomainNotActiveErrorType(t *testing.T) {
	assert.False(t, IsDomainNotActiveErrorType(errors.New("a random error")))
	assert.True(t, IsDomainNotActiveErrorType(&types.DomainNotActiveError{Message: "test"}))
}

func TestWorkflowExecutionAlreadyStartedErrorType(t *testing.T) {
	assert.False(t, IsWorkflowExecutionAlreadyStartedErrorType(errors.New("a random error")))
	assert.True(t, IsWorkflowExecutionAlreadyStartedErrorType(&types.WorkflowExecutionAlreadyStartedError{Message: "test"}))
}

func TestEntityNotExistsErrorType(t *testing.T) {
	assert.False(t, IsEntityNotExistsErrorType(errors.New("a random error")))
	assert.True(t, IsEntityNotExistsErrorType(&types.EntityNotExistsError{Message: "test"}))
}

func TestWorkflowExecutionAlreadyCompletedErrorType(t *testing.T) {
	assert.False(t, IsWorkflowExecutionAlreadyCompletedErrorType(errors.New("a random error")))
	assert.True(t, IsWorkflowExecutionAlreadyCompletedErrorType(&types.WorkflowExecutionAlreadyCompletedError{Message: "test"}))
}

func TestCancellationAlreadyRequestedErrorType(t *testing.T) {
	assert.False(t, IsCancellationAlreadyRequestedErrorType(errors.New("a random error")))
	assert.True(t, IsCancellationAlreadyRequestedErrorType(&types.CancellationAlreadyRequestedError{Message: "test"}))
}

func TestLimitExceededErrorType(t *testing.T) {
	assert.False(t, IsLimitExceededErrorType(errors.New("a random error")))
	assert.True(t, IsLimitExceededErrorType(&types.LimitExceededError{Message: "test"}))
}

func TestRetryTaskV2ErrorType(t *testing.T) {
	assert.False(t, IsRetryTaskV2ErrorType(errors.New("a random error")))
	assert.True(t, IsRetryTaskV2ErrorType(&types.RetryTaskV2Error{Message: "test"}))
}

func TestServiceBusyErrorType(t *testing.T) {
	assert.False(t, IsServiceBusyErrorType(errors.New("a random error")))
	assert.True(t, IsServiceBusyErrorType(&types.ServiceBusyError{Message: "test"}))
}

package constants

import "github.com/uber/cadence/common/types"

var (
	ErrDomainNotSet            = &types.BadRequestError{Message: "Domain not set on request."}
	ErrWorkflowExecutionNotSet = &types.BadRequestError{Message: "WorkflowExecution not set on request."}
	ErrTaskListNotSet          = &types.BadRequestError{Message: "Tasklist not set."}
	ErrRunIDNotValid           = &types.BadRequestError{Message: "RunID is not valid UUID."}
	ErrWorkflowIDNotSet        = &types.BadRequestError{Message: "WorkflowId is not set on request."}
	ErrSourceClusterNotSet     = &types.BadRequestError{Message: "Source Cluster not set on request."}
	ErrTimestampNotSet         = &types.BadRequestError{Message: "Timestamp not set on request."}
	ErrInvalidTaskType         = &types.BadRequestError{Message: "Invalid task type"}
	ErrHistoryHostThrottle     = &types.ServiceBusyError{Message: "History host rps exceeded"}
	ErrShuttingDown            = &types.InternalServiceError{Message: "Shutting down"}
)

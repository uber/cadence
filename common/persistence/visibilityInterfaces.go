package persistence

import (
	"time"

	s "github.com/uber/cadence/.gen/go/shared"
)

// Interfaces for the Visibility Store.
// This is a secondary store that is eventually consistent with the main
// executions store, and stores workflow execution records for visibility
// purposes.

type (
	// RecordWorkflowExecutionStartedRequest is used to add a record of a newly
	// started execution
	RecordWorkflowExecutionStartedRequest struct {
		// TODO: add domain info
		Execution        s.WorkflowExecution
		WorkflowTypeName string
		StartTimestamp   time.Time
	}

	// RecordWorkflowExecutionClosedRequest is used to add a record of a newly
	// closed execution
	RecordWorkflowExecutionClosedRequest struct {
		// TODO: add domain info
		Execution        s.WorkflowExecution
		WorkflowTypeName string
		StartTimestamp   time.Time
		CloseTimestamp   time.Time
	}

	// VisibilityManager is used to manage the visibility store
	VisibilityManager interface {
		RecordWorkflowExecutionStarted(request *RecordWorkflowExecutionStartedRequest) error
		RecordWorkflowExecutionClosed(request *RecordWorkflowExecutionClosedRequest) error
		// TODO: List APIs
	}
)

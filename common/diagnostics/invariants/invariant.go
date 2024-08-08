package invariants

import "context"

// InvariantCheckResult is the result from the invariant check
type InvariantCheckResult struct {
	InvariantType string
	Reason        string
	Metadata      []byte
}

// Invariant represents a condition of a workflow execution.
type Invariant interface {
	Check(context.Context) ([]InvariantCheckResult, error)
}

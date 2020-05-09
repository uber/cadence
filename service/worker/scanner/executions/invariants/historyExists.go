package invariants

import "github.com/uber/cadence/service/worker/scanner/executions/common"

type (
	historyExists struct {
		// add everything you need to do this check

	}
)

func NewHistoryExists() common.Invariant {
	return &historyExists{}
}

func (h *historyExists) Check(execution common.Execution, _ *common.InvariantResourceBag) common.CheckResult {
	return common.CheckResult{}
}

func (h *historyExists) Fix(execution common.Execution, _ *common.InvariantResourceBag) common.FixResult {
	return common.FixResult{}
}

func (h *historyExists) InvariantType() common.InvariantType {
	return common.HistoryExistsInvariantType
}

/**
// Invariant represents an invariant of a single execution.
// It can be used to check that the execution satisfies the invariant.
// It can also be used to fix the invariant for an execution.
Invariant interface {
	Check(Execution, *InvariantResourceBag) CheckResult
	Fix(Execution, *InvariantResourceBag) FixResult
	InvariantType() InvariantType
}
 */
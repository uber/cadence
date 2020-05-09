package invariants

import (
	c "github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

const (
	historyPageSize = 1
)

type (
	historyExists struct {
		pr common.PersistenceRetryer
	}
)

func NewHistoryExists(
	pr common.PersistenceRetryer,
	) common.Invariant {
	return &historyExists{
		pr: pr,
	}
}

func (h *historyExists) Check(execution common.Execution, _ *common.InvariantResourceBag) common.CheckResult {
	readHistoryBranchReq := &persistence.ReadHistoryBranchRequest{
		BranchToken: execution.BranchToken,
		MinEventID: c.FirstEventID,
		MaxEventID: c.EndEventID,
		PageSize: historyPageSize,
		NextPageToken: nil,
		ShardID: c.IntPtr(execution.ShardID),
	}
	readHistoryBranchResp, ReadHistoryBranchErr := h.pr.ReadHistoryBranch(readHistoryBranchReq)






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
package invariants

import (
	"github.com/gocql/gocql"

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

// NewHistoryExists returns a new history exists invariant
func NewHistoryExists(
	pr common.PersistenceRetryer,
) common.Invariant {
	return &historyExists{
		pr: pr,
	}
}

func (h *historyExists) Check(execution common.Execution, resources *common.InvariantResourceBag) common.CheckResult {
	readHistoryBranchReq := &persistence.ReadHistoryBranchRequest{
		BranchToken:   execution.BranchToken,
		MinEventID:    c.FirstEventID,
		MaxEventID:    c.EndEventID,
		PageSize:      historyPageSize,
		NextPageToken: nil,
		ShardID:       c.IntPtr(execution.ShardID),
	}
	readHistoryBranchResp, readHistoryBranchErr := h.pr.ReadHistoryBranch(readHistoryBranchReq)
	stillExists, existsCheckError := common.ExecutionStillExists(&execution, h.pr)
	if existsCheckError != nil {
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeFailed,
			Info:            "failed to check if concrete execution still exists",
			InfoDetails:     existsCheckError.Error(),
		}
	}
	if !stillExists {
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeHealthy,
			Info:            "determined execution was healthy because concrete execution no longer exists",
		}
	}
	if readHistoryBranchErr != nil {
		if readHistoryBranchErr == gocql.ErrNotFound {
			return common.CheckResult{
				CheckResultType: common.CheckResultTypeCorrupted,
				Info:            "concrete execution exists but history does not exist",
				InfoDetails:     readHistoryBranchErr.Error(),
			}
		}
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeFailed,
			Info:            "failed to verify if history exists",
			InfoDetails:     readHistoryBranchErr.Error(),
		}
	}
	if readHistoryBranchResp == nil || len(readHistoryBranchResp.HistoryEvents) == 0 {
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeCorrupted,
			Info:            "concrete execution exists but got empty history",
		}
	}
	resources.History = readHistoryBranchResp
	return common.CheckResult{
		CheckResultType: common.CheckResultTypeHealthy,
	}
}

func (h *historyExists) Fix(execution common.Execution) common.FixResult {
	return common.FixResult{}
}

func (h *historyExists) InvariantType() common.InvariantType {
	return common.HistoryExistsInvariantType
}

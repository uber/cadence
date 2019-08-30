package persistencetests

import (
	"github.com/uber/cadence/common/persistence"
	"testing"
)

func TestGarbageCleanupInfo(t *testing.T) {
	domainID := "10000000-5000-f000-f000-000000000000"
	workflowID := "workflow-id"
	runID := "10000000-5000-f000-f000-000000000002"

	info := persistence.BuildHistoryGarbageCleanupInfo(domainID, workflowID, runID)
	domainID2, workflowID2, runID2, err := persistence.SplitHistoryGarbageCleanupInfo(info)
	if err != nil || domainID != domainID2 || workflowID != workflowID2 || runID != runID2 {
		t.Fail()
	}
}

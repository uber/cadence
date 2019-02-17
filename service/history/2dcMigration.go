// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package history

import (
	"context"
	"github.com/uber-common/bark"
	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

// TODO remove after DC migration is over

type (
	dcMigrationHandler struct {
		historyMgr   persistence.HistoryManager
		historyV2Mgr persistence.HistoryV2Manager
	}
)

func newDCMigrationHandler(historyMgr persistence.HistoryManager, historyV2Mgr persistence.HistoryV2Manager) *dcMigrationHandler {
	return &dcMigrationHandler{
		historyMgr:   historyMgr,
		historyV2Mgr: historyV2Mgr,
	}
}

func (r *dcMigrationHandler) getLastMatchEventID(ctx context.Context, context workflowExecutionContext,
	msBuilder mutableState, request *h.ReplicateEventsRequest, logger bark.Logger) (int64, error) {

	// NOTE: this is a temporary solution for DC migration
	// DO NOT USE UNLESS YOU KNOW WHAT YOU ARE DOING

	var lastCommonEventID *int64
	for clusterName, localReplicationInfo := range msBuilder.GetReplicationState().LastReplicationInfo {

		if remoteReplicationInfo, ok := request.ReplicationInfo[clusterName]; ok {

			if remoteReplicationInfo.GetVersion() == localReplicationInfo.Version {

				if remoteReplicationInfo.GetLastEventId() < localReplicationInfo.LastEventID {

					if lastCommonEventID == nil || remoteReplicationInfo.GetLastEventId() >= *lastCommonEventID {
						lastCommonEventID = common.Int64Ptr(remoteReplicationInfo.GetLastEventId())
					}

				} else {

					if lastCommonEventID == nil || localReplicationInfo.LastEventID >= *lastCommonEventID {
						lastCommonEventID = common.Int64Ptr(localReplicationInfo.LastEventID)
					}

				}

			}

		}
	}
	if lastCommonEventID == nil {
		// reset to 1 batch
		executionInfo := msBuilder.GetExecutionInfo()
		return r.getFirstBatchLastEventID(context.getDomainID(), context.getExecution(), executionInfo.EventStoreVersion, executionInfo.GetCurrentBranch())
	}

	return *lastCommonEventID, nil
}

func (r *dcMigrationHandler) getFirstBatchLastEventID(domainID string, execution *shared.WorkflowExecution, eventStoreVersion int32, branchToken []byte) (int64, error) {

	if eventStoreVersion == persistence.EventStoreVersionV2 {
		response, err := r.historyV2Mgr.ReadHistoryBranch(&persistence.ReadHistoryBranchRequest{
			BranchToken:   branchToken,
			MinEventID:    common.FirstEventID,
			MaxEventID:    common.FirstEventID + 1,
			PageSize:      1,
			NextPageToken: nil,
		})
		if err != nil {
			return 0, err
		}
		return response.HistoryEvents[len(response.HistoryEvents)-1].GetEventId(), nil
	}

	response, err := r.historyMgr.GetWorkflowExecutionHistory(&persistence.GetWorkflowExecutionHistoryRequest{
		DomainID:      domainID,
		Execution:     *execution,
		FirstEventID:  common.FirstEventID,
		NextEventID:   common.FirstEventID + 1,
		PageSize:      1,
		NextPageToken: nil,
	})

	if err != nil {
		return 0, err
	}
	return response.History.Events[len(response.History.Events)-1].GetEventId(), nil
}

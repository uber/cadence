// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testdata

import (
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

const (
	TaskState = int16(1)
)

var (
	CrossClusterTaskInfo                             = *generateCrossClusterTaskInfo(types.CrossClusterTaskTypeStartChildExecution)
	CrossClusterStartChildExecutionRequestAttributes = types.CrossClusterStartChildExecutionRequestAttributes{
		TargetDomainID:           DomainID,
		RequestID:                RequestID,
		InitiatedEventID:         EventID1,
		InitiatedEventAttributes: &StartChildWorkflowExecutionInitiatedEventAttributes,
		TargetRunID:              common.StringPtr(RunID1),
	}
	CrossClusterStartChildExecutionResponseAttributes = types.CrossClusterStartChildExecutionResponseAttributes{
		RunID: RunID,
	}
	CrossClusterCancelExecutionRequestAttributes = types.CrossClusterCancelExecutionRequestAttributes{
		TargetDomainID:    DomainID,
		TargetWorkflowID:  WorkflowID,
		TargetRunID:       RunID,
		RequestID:         RequestID,
		InitiatedEventID:  EventID1,
		ChildWorkflowOnly: true,
	}
	CrossClusterCancelExecutionResponseAttributes = types.CrossClusterCancelExecutionResponseAttributes{}
	CrossClusterSignalExecutionRequestAttributes  = types.CrossClusterSignalExecutionRequestAttributes{
		TargetDomainID:    DomainID,
		TargetWorkflowID:  WorkflowID,
		TargetRunID:       RunID,
		RequestID:         RequestID,
		InitiatedEventID:  EventID1,
		ChildWorkflowOnly: true,
		SignalName:        SignalName,
		SignalInput:       Payload1,
		Control:           Control,
	}
	CrossClusterSignalExecutionResponseAttributes                     = types.CrossClusterSignalExecutionResponseAttributes{}
	CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes = types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes{
		TargetDomainID:   DomainID,
		TargetWorkflowID: WorkflowID,
		TargetRunID:      RunID,
		InitiatedEventID: EventID1,
		CompletionEvent:  &HistoryEvent_ChildWorkflowExecutionCompleted,
	}

	CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes = types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes{}
	CrossClusterApplyParentClosePolicyRequestAttributes                = types.CrossClusterApplyParentClosePolicyRequestAttributes{
		Children: []*types.ApplyParentClosePolicyRequest{
			{
				Child: &types.ApplyParentClosePolicyAttributes{
					ChildDomainID:     DomainID,
					ChildWorkflowID:   WorkflowID,
					ChildRunID:        RunID,
					ParentClosePolicy: &ParentClosePolicy,
				},
				Status: &types.ApplyParentClosePolicyStatus{
					Completed:   false,
					FailedCause: types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(),
				},
			},
		},
	}
	CrossClusterApplyParentClosePolicyResponseAttributes = types.CrossClusterApplyParentClosePolicyResponseAttributes{}
	CrossClusterTaskRequestStartChildExecution           = types.CrossClusterTaskRequest{
		TaskInfo:                      generateCrossClusterTaskInfo(types.CrossClusterTaskTypeStartChildExecution),
		StartChildExecutionAttributes: &CrossClusterStartChildExecutionRequestAttributes,
	}
	CrossClusterTaskRequestCancelExecution = types.CrossClusterTaskRequest{
		TaskInfo:                  generateCrossClusterTaskInfo(types.CrossClusterTaskTypeCancelExecution),
		CancelExecutionAttributes: &CrossClusterCancelExecutionRequestAttributes,
	}
	CrossClusterTaskRequestSignalExecution = types.CrossClusterTaskRequest{
		TaskInfo:                  generateCrossClusterTaskInfo(types.CrossClusterTaskTypeSignalExecution),
		SignalExecutionAttributes: &CrossClusterSignalExecutionRequestAttributes,
	}
	CrossClusterTaskRequestRecordChildExecutionComplete = types.CrossClusterTaskRequest{
		TaskInfo: generateCrossClusterTaskInfo(types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete),
		RecordChildWorkflowExecutionCompleteAttributes: &CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes,
	}
	CrossClusterTaskRequestApplyParentClosePolicy = types.CrossClusterTaskRequest{
		TaskInfo:                         generateCrossClusterTaskInfo(types.CrossClusterTaskTypeApplyParentPolicy),
		ApplyParentClosePolicyAttributes: &CrossClusterApplyParentClosePolicyRequestAttributes,
	}
	CrossClusterTaskResponseStartChildExecution = types.CrossClusterTaskResponse{
		TaskID:                        TaskID,
		TaskType:                      types.CrossClusterTaskTypeStartChildExecution.Ptr(),
		TaskState:                     1,
		StartChildExecutionAttributes: &CrossClusterStartChildExecutionResponseAttributes,
	}
	CrossClusterTaskResponseCancelExecution = types.CrossClusterTaskResponse{
		TaskID:      TaskID,
		TaskType:    types.CrossClusterTaskTypeCancelExecution.Ptr(),
		TaskState:   3,
		FailedCause: types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(),
	}
	CrossClusterTaskResponseSignalExecution = types.CrossClusterTaskResponse{
		TaskID:      TaskID,
		TaskType:    types.CrossClusterTaskTypeSignalExecution.Ptr(),
		TaskState:   3,
		FailedCause: types.CrossClusterTaskFailedCauseWorkflowNotExists.Ptr(),
	}
	CrossClusterTaskResponseRecordChildExecutionComplete = types.CrossClusterTaskResponse{
		TaskID:    TaskID,
		TaskType:  types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete.Ptr(),
		TaskState: 1,
		RecordChildWorkflowExecutionCompleteAttributes: &CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes,
	}
	CrossClusterTaskResponseApplyParentClosePolicy = types.CrossClusterTaskResponse{
		TaskID:                           TaskID,
		TaskType:                         types.CrossClusterTaskTypeApplyParentPolicy.Ptr(),
		TaskState:                        1,
		ApplyParentClosePolicyAttributes: &CrossClusterApplyParentClosePolicyResponseAttributes,
	}
	CrossClusterTaskRequestArray = []*types.CrossClusterTaskRequest{
		&CrossClusterTaskRequestStartChildExecution,
		&CrossClusterTaskRequestCancelExecution,
		&CrossClusterTaskRequestSignalExecution,
		&CrossClusterTaskRequestRecordChildExecutionComplete,
		&CrossClusterTaskRequestApplyParentClosePolicy,
	}
	CrossClusterTaskResponseArray = []*types.CrossClusterTaskResponse{
		&CrossClusterTaskResponseStartChildExecution,
		&CrossClusterTaskResponseCancelExecution,
		&CrossClusterTaskResponseSignalExecution,
		&CrossClusterTaskResponseRecordChildExecutionComplete,
		&CrossClusterTaskResponseApplyParentClosePolicy,
	}
	CrossClusterTaskRequestMap = map[int32][]*types.CrossClusterTaskRequest{
		ShardID + 1: {},
		ShardID + 2: CrossClusterTaskRequestArray,
	}
	GetCrossClusterTaskFailedCauseMap = map[int32]types.GetTaskFailedCause{
		ShardID + 3: types.GetTaskFailedCauseServiceBusy,
		ShardID + 4: types.GetTaskFailedCauseTimeout,
		ShardID + 5: types.GetTaskFailedCauseShardOwnershipLost,
		ShardID + 6: types.GetTaskFailedCauseUncategorized,
	}
	GetCrossClusterTasksRequest = types.GetCrossClusterTasksRequest{
		ShardIDs:      []int32{ShardID},
		TargetCluster: ClusterName1,
	}
	GetCrossClusterTasksResponse = types.GetCrossClusterTasksResponse{
		TasksByShard:       CrossClusterTaskRequestMap,
		FailedCauseByShard: GetCrossClusterTaskFailedCauseMap,
	}
	RespondCrossClusterTasksCompletedRequest = types.RespondCrossClusterTasksCompletedRequest{
		ShardID:       ShardID,
		TargetCluster: ClusterName1,
		TaskResponses: CrossClusterTaskResponseArray,
		FetchNewTasks: true,
	}
	RespondCrossClusterTasksCompletedResponse = types.RespondCrossClusterTasksCompletedResponse{
		Tasks: CrossClusterTaskRequestArray,
	}
	ApplyParentClosePolicyAttributes = types.ApplyParentClosePolicyAttributes{
		ChildDomainID:     "testDomain",
		ChildWorkflowID:   "testChildWorkflowID",
		ChildRunID:        "testChildRunID",
		ParentClosePolicy: &ParentClosePolicy,
	}
	ApplyParentClosePolicyResult = types.ApplyParentClosePolicyResult{
		Child:       &ApplyParentClosePolicyAttributes,
		FailedCause: types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(),
	}
	ApplyParentClosePolicyResult2 = types.ApplyParentClosePolicyResult{
		Child: &types.ApplyParentClosePolicyAttributes{
			ChildDomainID:     "testDomain2",
			ChildWorkflowID:   "testChildWorkflowID2",
			ChildRunID:        "testChildRunID2",
			ParentClosePolicy: &ParentClosePolicy2,
		},
		FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr(),
	}
	CrossClusterApplyParentClosePolicyResponseWithChildren = types.CrossClusterApplyParentClosePolicyResponseAttributes{
		ChildrenStatus: []*types.ApplyParentClosePolicyResult{
			&ApplyParentClosePolicyResult,
			&ApplyParentClosePolicyResult2,
		},
	}
)

func generateCrossClusterTaskInfo(
	taskType types.CrossClusterTaskType,
) *types.CrossClusterTaskInfo {
	return &types.CrossClusterTaskInfo{
		DomainID:            DomainID,
		WorkflowID:          WorkflowID,
		RunID:               RunID,
		TaskType:            taskType.Ptr(),
		TaskState:           TaskState,
		VisibilityTimestamp: &Timestamp,
	}
}

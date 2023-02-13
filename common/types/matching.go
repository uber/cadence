// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package types

import (
	"fmt"
	"strconv"
	"strings"
)

// AddActivityTaskRequest is an internal type (TBD...)
type AddActivityTaskRequest struct {
	DomainUUID                    string
	Execution                     *WorkflowExecution
	SourceDomainUUID              string
	TaskList                      *TaskList
	ScheduleID                    int64
	ScheduleToStartTimeoutSeconds *int32
	Source                        *TaskSource
	ForwardedFrom                 string
	ActivityTaskDispatchInfo      *ActivityTaskDispatchInfo
}

// GetDomainUUID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetSourceDomainUUID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetSourceDomainUUID() (o string) {
	if v != nil {
		return v.SourceDomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetScheduleID() (o int64) {
	if v != nil {
		return v.ScheduleID
	}
	return
}

// GetScheduleToStartTimeoutSeconds is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

// GetSource is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetSource() (o TaskSource) {
	if v != nil && v.Source != nil {
		return *v.Source
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetForwardedFrom() (o string) {
	if v != nil {
		return v.ForwardedFrom
	}
	return
}

// ActivityTaskDispatchInfo is an internal type (TBD...)
type ActivityTaskDispatchInfo struct {
	ScheduledEvent                  *HistoryEvent
	StartedTimestamp                *int64
	Attempt                         *int64
	ScheduledTimestampOfThisAttempt *int64
	HeartbeatDetails                []byte
	WorkflowType                    *WorkflowType
	WorkflowDomain                  string
}

// AddDecisionTaskRequest is an internal type (TBD...)
type AddDecisionTaskRequest struct {
	DomainUUID                    string
	Execution                     *WorkflowExecution
	TaskList                      *TaskList
	ScheduleID                    int64
	ScheduleToStartTimeoutSeconds *int32
	Source                        *TaskSource
	ForwardedFrom                 string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetScheduleID() (o int64) {
	if v != nil {
		return v.ScheduleID
	}
	return
}

// GetScheduleToStartTimeoutSeconds is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

// GetSource is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetSource() (o TaskSource) {
	if v != nil && v.Source != nil {
		return *v.Source
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetForwardedFrom() (o string) {
	if v != nil {
		return v.ForwardedFrom
	}
	return
}

// CancelOutstandingPollRequest is an internal type (TBD...)
type CancelOutstandingPollRequest struct {
	DomainUUID   string
	TaskListType *int32
	TaskList     *TaskList
	PollerID     string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetTaskListType is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetTaskListType() (o int32) {
	if v != nil && v.TaskListType != nil {
		return *v.TaskListType
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetPollerID() (o string) {
	if v != nil {
		return v.PollerID
	}
	return
}

// MatchingDescribeTaskListRequest is an internal type (TBD...)
type MatchingDescribeTaskListRequest struct {
	DomainUUID  string
	DescRequest *DescribeTaskListRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingDescribeTaskListRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetDescRequest is an internal getter (TBD...)
func (v *MatchingDescribeTaskListRequest) GetDescRequest() (o *DescribeTaskListRequest) {
	if v != nil && v.DescRequest != nil {
		return v.DescRequest
	}
	return
}

// MatchingListTaskListPartitionsRequest is an internal type (TBD...)
type MatchingListTaskListPartitionsRequest struct {
	Domain   string
	TaskList *TaskList
}

// GetDomain is an internal getter (TBD...)
func (v *MatchingListTaskListPartitionsRequest) GetDomain() (o string) {
	if v != nil {
		return v.Domain
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingListTaskListPartitionsRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// MatchingGetTaskListsByDomainRequest is an internal type (TBD...)
type MatchingGetTaskListsByDomainRequest struct {
	Domain string
}

// GetDomainName is an internal getter (TBD...)
func (v *MatchingGetTaskListsByDomainRequest) GetDomain() (o string) {
	if v != nil {
		return v.Domain
	}
	return
}

// MatchingPollForActivityTaskRequest is an internal type (TBD...)
type MatchingPollForActivityTaskRequest struct {
	DomainUUID    string
	PollerID      string
	PollRequest   *PollForActivityTaskRequest
	ForwardedFrom string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetPollerID() (o string) {
	if v != nil {
		return v.PollerID
	}
	return
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetPollRequest() (o *PollForActivityTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetForwardedFrom() (o string) {
	if v != nil {
		return v.ForwardedFrom
	}
	return
}

// MatchingPollForDecisionTaskRequest is an internal type (TBD...)
type MatchingPollForDecisionTaskRequest struct {
	DomainUUID    string
	PollerID      string
	PollRequest   *PollForDecisionTaskRequest
	ForwardedFrom string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetPollerID() (o string) {
	if v != nil {
		return v.PollerID
	}
	return
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetPollRequest() (o *PollForDecisionTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetForwardedFrom() (o string) {
	if v != nil {
		return v.ForwardedFrom
	}
	return
}

// MatchingPollForDecisionTaskResponse is an internal type (TBD...)
type MatchingPollForDecisionTaskResponse struct {
	TaskToken                 []byte
	WorkflowExecution         *WorkflowExecution
	WorkflowType              *WorkflowType
	PreviousStartedEventID    *int64
	StartedEventID            int64
	Attempt                   int64
	NextEventID               int64
	BacklogCountHint          int64
	StickyExecutionEnabled    bool
	Query                     *WorkflowQuery
	DecisionInfo              *TransientDecisionInfo
	WorkflowExecutionTaskList *TaskList
	EventStoreVersion         int32
	BranchToken               []byte
	ScheduledTimestamp        *int64
	StartedTimestamp          *int64
	Queries                   map[string]*WorkflowQuery
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

// GetPreviousStartedEventID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetPreviousStartedEventID() (o int64) {
	if v != nil && v.PreviousStartedEventID != nil {
		return *v.PreviousStartedEventID
	}
	return
}

// GetNextEventID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetNextEventID() (o int64) {
	if v != nil {
		return v.NextEventID
	}
	return
}

// GetStickyExecutionEnabled is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetStickyExecutionEnabled() (o bool) {
	if v != nil {
		return v.StickyExecutionEnabled
	}
	return
}

// GetBranchToken is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetBranchToken() (o []byte) {
	if v != nil && v.BranchToken != nil {
		return v.BranchToken
	}
	return
}

// MatchingQueryWorkflowRequest is an internal type (TBD...)
type MatchingQueryWorkflowRequest struct {
	DomainUUID    string
	TaskList      *TaskList
	QueryRequest  *QueryWorkflowRequest
	ForwardedFrom string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetQueryRequest is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetQueryRequest() (o *QueryWorkflowRequest) {
	if v != nil && v.QueryRequest != nil {
		return v.QueryRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetForwardedFrom() (o string) {
	if v != nil {
		return v.ForwardedFrom
	}
	return
}

// MatchingRespondQueryTaskCompletedRequest is an internal type (TBD...)
type MatchingRespondQueryTaskCompletedRequest struct {
	DomainUUID       string
	TaskList         *TaskList
	TaskID           string
	CompletedRequest *RespondQueryTaskCompletedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetTaskID is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetTaskID() (o string) {
	if v != nil {
		return v.TaskID
	}
	return
}

// GetCompletedRequest is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetCompletedRequest() (o *RespondQueryTaskCompletedRequest) {
	if v != nil && v.CompletedRequest != nil {
		return v.CompletedRequest
	}
	return
}

// TaskSource is an internal type (TBD...)
type TaskSource int32

// Ptr is a helper function for getting pointer value
func (e TaskSource) Ptr() *TaskSource {
	return &e
}

// String returns a readable string representation of TaskSource.
func (e TaskSource) String() string {
	w := int32(e)
	switch w {
	case 0:
		return "HISTORY"
	case 1:
		return "DB_BACKLOG"
	}
	return fmt.Sprintf("TaskSource(%d)", w)
}

// UnmarshalText parses enum value from string representation
func (e *TaskSource) UnmarshalText(value []byte) error {
	switch s := strings.ToUpper(string(value)); s {
	case "HISTORY":
		*e = TaskSourceHistory
		return nil
	case "DB_BACKLOG":
		*e = TaskSourceDbBacklog
		return nil
	default:
		val, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return fmt.Errorf("unknown enum value %q for %q: %v", s, "TaskSource", err)
		}
		*e = TaskSource(val)
		return nil
	}
}

// MarshalText encodes TaskSource to text.
func (e TaskSource) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

const (
	// TaskSourceHistory is an option for TaskSource
	TaskSourceHistory TaskSource = iota
	// TaskSourceDbBacklog is an option for TaskSource
	TaskSourceDbBacklog
)

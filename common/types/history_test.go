// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	domainUUID             = "test-domainUUID"
	mutableStateInCache    = "test-cache"
	mutableStateInDatabase = "test-database"
	domainIDs              = []string{"id1", "id2", "id3"}
	shardIDs               = []int32{1, 2, 3}
	expectedNextEventID    = int64(123)
	nextEventID            = int64(456)
	previousStartedEventID = int64(789)
	stickyTaskListTimeout  = int32(100)
	workflowCloseState     = int32(2)
	clientFeatureVersion   = "v1.0"
	clientImpl             = "test-client"
	currentBranchToken     = []byte("branch-token")
)

func TestDescribeMutableStateRequest(t *testing.T) {
	testStruct := DescribeMutableStateRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *DescribeMutableStateRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res) // Should return empty string when struct is nil
}

func TestDescribeMutableStateResponse(t *testing.T) {
	testStruct := DescribeMutableStateResponse{
		MutableStateInCache:    mutableStateInCache,
		MutableStateInDatabase: mutableStateInDatabase,
	}

	assert.Equal(t, mutableStateInCache, testStruct.MutableStateInCache)
	assert.Equal(t, mutableStateInDatabase, testStruct.MutableStateInDatabase)
}

func TestHistoryDescribeWorkflowExecutionRequest_GetDomainUUID(t *testing.T) {
	testStruct := HistoryDescribeWorkflowExecutionRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *HistoryDescribeWorkflowExecutionRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res) // Should return empty string when struct is nil
}

func TestDomainFilter_GetDomainIDs(t *testing.T) {
	testStruct := DomainFilter{
		DomainIDs: domainIDs,
	}

	// Non-nil struct test
	res := testStruct.GetDomainIDs()
	assert.Equal(t, domainIDs, res)

	// Nil struct test
	var nilStruct *DomainFilter
	res = nilStruct.GetDomainIDs()
	assert.Empty(t, res) // Should return empty slice when struct is nil
}

func TestDomainFilter_GetReverseMatch(t *testing.T) {
	testStruct := DomainFilter{
		ReverseMatch: true,
	}

	// Non-nil struct test
	res := testStruct.GetReverseMatch()
	assert.True(t, res)

	// Nil struct test
	var nilStruct *DomainFilter
	res = nilStruct.GetReverseMatch()
	assert.False(t, res) // Should return false when struct is nil
}

func TestFailoverMarkerToken_GetShardIDs(t *testing.T) {
	testStruct := FailoverMarkerToken{
		ShardIDs: shardIDs,
	}

	// Non-nil struct test
	res := testStruct.GetShardIDs()
	assert.Equal(t, shardIDs, res)

	// Nil struct test
	var nilStruct *FailoverMarkerToken
	res = nilStruct.GetShardIDs()
	assert.Empty(t, res) // Should return empty slice when struct is nil
}

func TestFailoverMarkerToken_GetFailoverMarker(t *testing.T) {
	failoverMarker := &FailoverMarkerAttributes{}

	testStruct := FailoverMarkerToken{
		FailoverMarker: failoverMarker,
	}

	// Non-nil struct test
	res := testStruct.GetFailoverMarker()
	assert.Equal(t, failoverMarker, res)

	// Nil struct test
	var nilStruct *FailoverMarkerToken
	res = nilStruct.GetFailoverMarker()
	assert.Nil(t, res) // Should return nil when struct is nil
}

func TestGetMutableStateRequest_GetDomainUUID(t *testing.T) {
	testStruct := GetMutableStateRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *GetMutableStateRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res)
}

func TestGetMutableStateRequest_GetExpectedNextEventID(t *testing.T) {
	testStruct := GetMutableStateRequest{
		ExpectedNextEventID: expectedNextEventID,
	}

	// Non-nil struct test
	res := testStruct.GetExpectedNextEventID()
	assert.Equal(t, expectedNextEventID, res)

	// Nil struct test
	var nilStruct *GetMutableStateRequest
	res = nilStruct.GetExpectedNextEventID()
	assert.Equal(t, int64(0), res)
}

func TestGetMutableStateResponse_GetNextEventID(t *testing.T) {
	testStruct := GetMutableStateResponse{
		NextEventID: nextEventID,
	}

	// Non-nil struct test
	res := testStruct.GetNextEventID()
	assert.Equal(t, nextEventID, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetNextEventID()
	assert.Equal(t, int64(0), res)
}

func TestGetMutableStateResponse_GetPreviousStartedEventID(t *testing.T) {
	testStruct := GetMutableStateResponse{
		PreviousStartedEventID: &previousStartedEventID,
	}

	// Non-nil struct test
	res := testStruct.GetPreviousStartedEventID()
	assert.Equal(t, previousStartedEventID, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetPreviousStartedEventID()
	assert.Equal(t, int64(0), res)
}

func TestGetMutableStateResponse_GetStickyTaskList(t *testing.T) {
	stickyTaskList := &TaskList{}
	testStruct := GetMutableStateResponse{
		StickyTaskList: stickyTaskList,
	}

	// Non-nil struct test
	res := testStruct.GetStickyTaskList()
	assert.Equal(t, stickyTaskList, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetStickyTaskList()
	assert.Nil(t, res)
}

func TestGetMutableStateResponse_GetClientFeatureVersion(t *testing.T) {
	testStruct := GetMutableStateResponse{
		ClientFeatureVersion: clientFeatureVersion,
	}

	// Non-nil struct test
	res := testStruct.GetClientFeatureVersion()
	assert.Equal(t, clientFeatureVersion, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetClientFeatureVersion()
	assert.Equal(t, "", res)
}

func TestGetMutableStateResponse_GetClientImpl(t *testing.T) {
	testStruct := GetMutableStateResponse{
		ClientImpl: clientImpl,
	}

	// Non-nil struct test
	res := testStruct.GetClientImpl()
	assert.Equal(t, clientImpl, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetClientImpl()
	assert.Equal(t, "", res)
}

func TestGetMutableStateResponse_GetIsWorkflowRunning(t *testing.T) {
	testStruct := GetMutableStateResponse{
		IsWorkflowRunning: true,
	}

	// Non-nil struct test
	res := testStruct.GetIsWorkflowRunning()
	assert.True(t, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetIsWorkflowRunning()
	assert.False(t, res)
}

func TestGetMutableStateResponse_GetStickyTaskListScheduleToStartTimeout(t *testing.T) {
	testStruct := GetMutableStateResponse{
		StickyTaskListScheduleToStartTimeout: &stickyTaskListTimeout,
	}

	// Non-nil struct test
	res := testStruct.GetStickyTaskListScheduleToStartTimeout()
	assert.Equal(t, stickyTaskListTimeout, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetStickyTaskListScheduleToStartTimeout()
	assert.Equal(t, int32(0), res)
}

func TestGetMutableStateResponse_GetCurrentBranchToken(t *testing.T) {
	testStruct := GetMutableStateResponse{
		CurrentBranchToken: currentBranchToken,
	}

	// Non-nil struct test
	res := testStruct.GetCurrentBranchToken()
	assert.Equal(t, currentBranchToken, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetCurrentBranchToken()
	assert.Nil(t, res)
}

func TestGetMutableStateResponse_GetWorkflowCloseState(t *testing.T) {
	testStruct := GetMutableStateResponse{
		WorkflowCloseState: &workflowCloseState,
	}

	// Non-nil struct test
	res := testStruct.GetWorkflowCloseState()
	assert.Equal(t, workflowCloseState, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetWorkflowCloseState()
	assert.Equal(t, int32(0), res)
}

func TestGetMutableStateResponse_GetVersionHistories(t *testing.T) {
	versionHistories := &VersionHistories{}
	testStruct := GetMutableStateResponse{
		VersionHistories: versionHistories,
	}

	// Non-nil struct test
	res := testStruct.GetVersionHistories()
	assert.Equal(t, versionHistories, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetVersionHistories()
	assert.Nil(t, res)
}

func TestGetMutableStateResponse_GetIsStickyTaskListEnabled(t *testing.T) {
	testStruct := GetMutableStateResponse{
		IsStickyTaskListEnabled: true,
	}

	// Non-nil struct test
	res := testStruct.GetIsStickyTaskListEnabled()
	assert.True(t, res)

	// Nil struct test
	var nilStruct *GetMutableStateResponse
	res = nilStruct.GetIsStickyTaskListEnabled()
	assert.False(t, res)
}

func TestNotifyFailoverMarkersRequest_GetFailoverMarkerTokens(t *testing.T) {
	failoverMarkerTokens := []*FailoverMarkerToken{}
	testStruct := NotifyFailoverMarkersRequest{
		FailoverMarkerTokens: failoverMarkerTokens,
	}

	// Non-nil struct test
	res := testStruct.GetFailoverMarkerTokens()
	assert.Equal(t, failoverMarkerTokens, res)

	// Nil struct test
	var nilStruct *NotifyFailoverMarkersRequest
	res = nilStruct.GetFailoverMarkerTokens()
	assert.Nil(t, res)
}

func TestParentExecutionInfo_GetDomain(t *testing.T) {
	testStruct := ParentExecutionInfo{
		Domain: "test-domain",
	}

	// Non-nil struct test
	res := testStruct.GetDomain()
	assert.Equal(t, "test-domain", res)

	// Nil struct test
	var nilStruct *ParentExecutionInfo
	res = nilStruct.GetDomain()
	assert.Equal(t, "", res)
}

func TestParentExecutionInfo_GetExecution(t *testing.T) {
	execution := &WorkflowExecution{}
	testStruct := ParentExecutionInfo{
		Execution: execution,
	}

	// Non-nil struct test
	res := testStruct.GetExecution()
	assert.Equal(t, execution, res)

	// Nil struct test
	var nilStruct *ParentExecutionInfo
	res = nilStruct.GetExecution()
	assert.Nil(t, res)
}

func TestPollMutableStateRequest_GetDomainUUID(t *testing.T) {
	testStruct := PollMutableStateRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *PollMutableStateRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res)
}

func TestPollMutableStateResponse_GetNextEventID(t *testing.T) {
	testStruct := PollMutableStateResponse{
		NextEventID: nextEventID,
	}

	// Non-nil struct test
	res := testStruct.GetNextEventID()
	assert.Equal(t, nextEventID, res)

	// Nil struct test
	var nilStruct *PollMutableStateResponse
	res = nilStruct.GetNextEventID()
	assert.Equal(t, int64(0), res)
}

func TestPollMutableStateResponse_GetLastFirstEventID(t *testing.T) {
	testStruct := PollMutableStateResponse{
		LastFirstEventID: nextEventID,
	}

	// Non-nil struct test
	res := testStruct.GetLastFirstEventID()
	assert.Equal(t, nextEventID, res)

	// Nil struct test
	var nilStruct *PollMutableStateResponse
	res = nilStruct.GetLastFirstEventID()
	assert.Equal(t, int64(0), res)
}

func TestPollMutableStateResponse_GetWorkflowCloseState(t *testing.T) {
	testStruct := PollMutableStateResponse{
		WorkflowCloseState: &workflowCloseState,
	}

	// Non-nil struct test
	res := testStruct.GetWorkflowCloseState()
	assert.Equal(t, workflowCloseState, res)

	// Nil struct test
	var nilStruct *PollMutableStateResponse
	res = nilStruct.GetWorkflowCloseState()
	assert.Equal(t, int32(0), res)
}

func TestProcessingQueueState_GetLevel(t *testing.T) {
	level := int32(2)
	testStruct := ProcessingQueueState{
		Level: &level,
	}

	// Non-nil struct test
	res := testStruct.GetLevel()
	assert.Equal(t, level, res)

	// Nil struct test
	var nilStruct *ProcessingQueueState
	res = nilStruct.GetLevel()
	assert.Equal(t, int32(0), res)
}

func TestProcessingQueueState_GetAckLevel(t *testing.T) {
	ackLevel := int64(1111)
	testStruct := ProcessingQueueState{
		AckLevel: &ackLevel,
	}

	// Non-nil struct test
	res := testStruct.GetAckLevel()
	assert.Equal(t, ackLevel, res)

	// Nil struct test
	var nilStruct *ProcessingQueueState
	res = nilStruct.GetAckLevel()
	assert.Equal(t, int64(0), res)
}

func TestProcessingQueueState_GetMaxLevel(t *testing.T) {
	maxLevel := int64(2222)
	testStruct := ProcessingQueueState{
		MaxLevel: &maxLevel,
	}

	// Non-nil struct test
	res := testStruct.GetMaxLevel()
	assert.Equal(t, maxLevel, res)

	// Nil struct test
	var nilStruct *ProcessingQueueState
	res = nilStruct.GetMaxLevel()
	assert.Equal(t, int64(0), res)
}

func TestProcessingQueueState_GetDomainFilter(t *testing.T) {
	domainFilter := &DomainFilter{}
	testStruct := ProcessingQueueState{
		DomainFilter: domainFilter,
	}

	// Non-nil struct test
	res := testStruct.GetDomainFilter()
	assert.Equal(t, domainFilter, res)

	// Nil struct test
	var nilStruct *ProcessingQueueState
	res = nilStruct.GetDomainFilter()
	assert.Nil(t, res)
}

func TestHistoryQueryWorkflowRequest_GetDomainUUID(t *testing.T) {
	testStruct := HistoryQueryWorkflowRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *HistoryQueryWorkflowRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res)
}

func TestHistoryQueryWorkflowRequest_GetRequest(t *testing.T) {
	request := &QueryWorkflowRequest{}
	testStruct := HistoryQueryWorkflowRequest{
		Request: request,
	}

	// Non-nil struct test
	res := testStruct.GetRequest()
	assert.Equal(t, request, res)

	// Nil struct test
	var nilStruct *HistoryQueryWorkflowRequest
	res = nilStruct.GetRequest()
	assert.Nil(t, res)
}

func TestHistoryQueryWorkflowResponse_GetResponse(t *testing.T) {
	response := &QueryWorkflowResponse{}
	testStruct := HistoryQueryWorkflowResponse{
		Response: response,
	}

	// Non-nil struct test
	res := testStruct.GetResponse()
	assert.Equal(t, response, res)

	// Nil struct test
	var nilStruct *HistoryQueryWorkflowResponse
	res = nilStruct.GetResponse()
	assert.Nil(t, res)
}

func TestHistoryReapplyEventsRequest_GetDomainUUID(t *testing.T) {
	testStruct := HistoryReapplyEventsRequest{
		DomainUUID: domainUUID,
	}

	// Non-nil struct test
	res := testStruct.GetDomainUUID()
	assert.Equal(t, domainUUID, res)

	// Nil struct test
	var nilStruct *HistoryReapplyEventsRequest
	res = nilStruct.GetDomainUUID()
	assert.Equal(t, "", res)
}

func TestHistoryReapplyEventsRequest_GetRequest(t *testing.T) {
	request := &ReapplyEventsRequest{}
	testStruct := HistoryReapplyEventsRequest{
		Request: request,
	}

	// Non-nil struct test
	res := testStruct.GetRequest()
	assert.Equal(t, request, res)

	// Nil struct test
	var nilStruct *HistoryReapplyEventsRequest
	res = nilStruct.GetRequest()
	assert.Nil(t, res)
}

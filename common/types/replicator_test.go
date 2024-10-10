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

func TestDLQType_Ptr(t *testing.T) {
	dlqType := DLQTypeReplication
	ptr := dlqType.Ptr()

	assert.Equal(t, &dlqType, ptr)
}

func TestDLQType_String(t *testing.T) {
	dlqType := DLQTypeReplication
	assert.Equal(t, "Replication", dlqType.String())

	dlqType = DLQTypeDomain
	assert.Equal(t, "Domain", dlqType.String())

	dlqType = 2
	assert.Equal(t, "DLQType(2)", dlqType.String())
}

func TestDLQType_UnmarshalText(t *testing.T) {
	var dlqType DLQType
	err := dlqType.UnmarshalText([]byte("Replication"))
	assert.NoError(t, err)
	assert.Equal(t, DLQTypeReplication, dlqType)

	err = dlqType.UnmarshalText([]byte("Domain"))
	assert.NoError(t, err)
	assert.Equal(t, DLQTypeDomain, dlqType)

	err = dlqType.UnmarshalText([]byte("2"))
	assert.NoError(t, err)
	assert.Equal(t, DLQType(2), dlqType)

	err = dlqType.UnmarshalText([]byte("Invalid"))
	assert.Error(t, err)
}

func TestDLQType_MarshalText(t *testing.T) {
	dlqType := DLQTypeReplication
	text, err := dlqType.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Replication"), text)

	dlqType = DLQTypeDomain
	text, err = dlqType.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Domain"), text)
}

func TestDomainOperation_Ptr(t *testing.T) {
	domainOp := DomainOperationCreate
	ptr := domainOp.Ptr()

	assert.Equal(t, &domainOp, ptr)
}

func TestDomainOperation_String(t *testing.T) {
	domainOp := DomainOperationCreate
	assert.Equal(t, "Create", domainOp.String())

	domainOp = DomainOperationUpdate
	assert.Equal(t, "Update", domainOp.String())

	domainOp = 2
	assert.Equal(t, "DomainOperation(2)", domainOp.String())
}

func TestDomainOperation_UnmarshalText(t *testing.T) {
	var domainOp DomainOperation
	err := domainOp.UnmarshalText([]byte("Create"))
	assert.NoError(t, err)
	assert.Equal(t, DomainOperationCreate, domainOp)

	err = domainOp.UnmarshalText([]byte("Update"))
	assert.NoError(t, err)
	assert.Equal(t, DomainOperationUpdate, domainOp)

	err = domainOp.UnmarshalText([]byte("2"))
	assert.NoError(t, err)
	assert.Equal(t, DomainOperation(2), domainOp)

	err = domainOp.UnmarshalText([]byte("Invalid"))
	assert.Error(t, err)
}

func TestDomainOperation_MarshalText(t *testing.T) {
	domainOp := DomainOperationCreate
	text, err := domainOp.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Create"), text)

	domainOp = DomainOperationUpdate
	text, err = domainOp.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Update"), text)
}

func TestDomainTaskAttributes_GetDomainOperation(t *testing.T) {
	domainOp := DomainOperationCreate
	testStruct := DomainTaskAttributes{
		DomainOperation: &domainOp,
	}

	res := testStruct.GetDomainOperation()
	assert.Equal(t, domainOp, res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetDomainOperation()
	assert.Equal(t, DomainOperation(0), res)
}

func TestDomainTaskAttributes_GetID(t *testing.T) {
	testStruct := DomainTaskAttributes{
		ID: "test-id",
	}

	res := testStruct.GetID()
	assert.Equal(t, "test-id", res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetID()
	assert.Equal(t, "", res)
}

func TestDomainTaskAttributes_GetInfo(t *testing.T) {
	domainInfo := &DomainInfo{}
	testStruct := DomainTaskAttributes{
		Info: domainInfo,
	}

	res := testStruct.GetInfo()
	assert.Equal(t, domainInfo, res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetInfo()
	assert.Nil(t, res)
}

func TestDomainTaskAttributes_GetConfigVersion(t *testing.T) {
	testStruct := DomainTaskAttributes{
		ConfigVersion: 123,
	}

	res := testStruct.GetConfigVersion()
	assert.Equal(t, int64(123), res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetConfigVersion()
	assert.Equal(t, int64(0), res)
}

func TestDomainTaskAttributes_GetFailoverVersion(t *testing.T) {
	testStruct := DomainTaskAttributes{
		FailoverVersion: 456,
	}

	res := testStruct.GetFailoverVersion()
	assert.Equal(t, int64(456), res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetFailoverVersion()
	assert.Equal(t, int64(0), res)
}

func TestDomainTaskAttributes_GetPreviousFailoverVersion(t *testing.T) {
	testStruct := DomainTaskAttributes{
		PreviousFailoverVersion: 789,
	}

	res := testStruct.GetPreviousFailoverVersion()
	assert.Equal(t, int64(789), res)

	var nilStruct *DomainTaskAttributes
	res = nilStruct.GetPreviousFailoverVersion()
	assert.Equal(t, int64(0), res)
}

func TestFailoverMarkerAttributes_GetDomainID(t *testing.T) {
	testStruct := FailoverMarkerAttributes{
		DomainID: "domain-id",
	}

	res := testStruct.GetDomainID()
	assert.Equal(t, "domain-id", res)

	var nilStruct *FailoverMarkerAttributes
	res = nilStruct.GetDomainID()
	assert.Equal(t, "", res)
}

func TestFailoverMarkerAttributes_GetFailoverVersion(t *testing.T) {
	testStruct := FailoverMarkerAttributes{
		FailoverVersion: 1234,
	}

	res := testStruct.GetFailoverVersion()
	assert.Equal(t, int64(1234), res)

	var nilStruct *FailoverMarkerAttributes
	res = nilStruct.GetFailoverVersion()
	assert.Equal(t, int64(0), res)
}

func TestFailoverMarkerAttributes_GetCreationTime(t *testing.T) {
	creationTime := int64(5678)
	testStruct := FailoverMarkerAttributes{
		CreationTime: &creationTime,
	}

	res := testStruct.GetCreationTime()
	assert.Equal(t, creationTime, res)

	var nilStruct *FailoverMarkerAttributes
	res = nilStruct.GetCreationTime()
	assert.Equal(t, int64(0), res)
}

func TestMergeDLQMessagesRequest_GetType(t *testing.T) {
	dlqType := DLQTypeReplication
	testStruct := MergeDLQMessagesRequest{
		Type: &dlqType,
	}

	res := testStruct.GetType()
	assert.Equal(t, dlqType, res)

	var nilStruct *MergeDLQMessagesRequest
	res = nilStruct.GetType()
	assert.Equal(t, DLQType(0), res)
}

func TestMergeDLQMessagesRequest_GetShardID(t *testing.T) {
	testStruct := MergeDLQMessagesRequest{
		ShardID: 101,
	}

	res := testStruct.GetShardID()
	assert.Equal(t, int32(101), res)

	var nilStruct *MergeDLQMessagesRequest
	res = nilStruct.GetShardID()
	assert.Equal(t, int32(0), res)
}

func TestMergeDLQMessagesRequest_GetSourceCluster(t *testing.T) {
	testStruct := MergeDLQMessagesRequest{
		SourceCluster: "cluster-1",
	}

	res := testStruct.GetSourceCluster()
	assert.Equal(t, "cluster-1", res)

	var nilStruct *MergeDLQMessagesRequest
	res = nilStruct.GetSourceCluster()
	assert.Equal(t, "", res)
}

func TestMergeDLQMessagesRequest_GetInclusiveEndMessageID(t *testing.T) {
	endMessageID := int64(102)
	testStruct := MergeDLQMessagesRequest{
		InclusiveEndMessageID: &endMessageID,
	}

	res := testStruct.GetInclusiveEndMessageID()
	assert.Equal(t, endMessageID, res)

	var nilStruct *MergeDLQMessagesRequest
	res = nilStruct.GetInclusiveEndMessageID()
	assert.Equal(t, int64(0), res)
}

func TestGetDLQReplicationMessagesRequest_GetTaskInfos(t *testing.T) {
	taskInfos := []*ReplicationTaskInfo{{}, {}}
	testStruct := GetDLQReplicationMessagesRequest{
		TaskInfos: taskInfos,
	}

	res := testStruct.GetTaskInfos()
	assert.Equal(t, taskInfos, res)

	var nilStruct *GetDLQReplicationMessagesRequest
	res = nilStruct.GetTaskInfos()
	assert.Nil(t, res)
}

func TestGetDomainReplicationMessagesRequest_GetLastRetrievedMessageID(t *testing.T) {
	lastRetrievedMessageID := int64(12345)
	testStruct := GetDomainReplicationMessagesRequest{
		LastRetrievedMessageID: &lastRetrievedMessageID,
	}

	res := testStruct.GetLastRetrievedMessageID()
	assert.Equal(t, lastRetrievedMessageID, res)

	var nilStruct *GetDomainReplicationMessagesRequest
	res = nilStruct.GetLastRetrievedMessageID()
	assert.Equal(t, int64(0), res)
}

func TestGetDomainReplicationMessagesRequest_GetLastProcessedMessageID(t *testing.T) {
	lastProcessedMessageID := int64(67890)
	testStruct := GetDomainReplicationMessagesRequest{
		LastProcessedMessageID: &lastProcessedMessageID,
	}

	res := testStruct.GetLastProcessedMessageID()
	assert.Equal(t, lastProcessedMessageID, res)

	var nilStruct *GetDomainReplicationMessagesRequest
	res = nilStruct.GetLastProcessedMessageID()
	assert.Equal(t, int64(0), res)
}

func TestGetDomainReplicationMessagesRequest_GetClusterName(t *testing.T) {
	testStruct := GetDomainReplicationMessagesRequest{
		ClusterName: "test-cluster",
	}

	res := testStruct.GetClusterName()
	assert.Equal(t, "test-cluster", res)

	var nilStruct *GetDomainReplicationMessagesRequest
	res = nilStruct.GetClusterName()
	assert.Equal(t, "", res)
}

func TestGetReplicationMessagesRequest_SerializeForLogging(t *testing.T) {
	// Test case where the struct is nil
	var nilStruct *GetReplicationMessagesRequest
	res, err := nilStruct.SerializeForLogging()
	assert.Equal(t, "", res)
	assert.NoError(t, err)

	// Test case with a non-nil struct
	tokens := []*ReplicationToken{{}, {}}
	testStruct := GetReplicationMessagesRequest{
		Tokens:      tokens,
		ClusterName: "test-cluster",
	}

	res, err = testStruct.SerializeForLogging()
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}

func TestGetReplicationMessagesRequest_GetClusterName(t *testing.T) {
	testStruct := GetReplicationMessagesRequest{
		ClusterName: "test-cluster",
	}

	res := testStruct.GetClusterName()
	assert.Equal(t, "test-cluster", res)

	var nilStruct *GetReplicationMessagesRequest
	res = nilStruct.GetClusterName()
	assert.Equal(t, "", res)
}

func TestGetReplicationMessagesResponse_GetMessagesByShard(t *testing.T) {
	messagesByShard := map[int32]*ReplicationMessages{
		1: {},
		2: {},
	}
	testStruct := GetReplicationMessagesResponse{
		MessagesByShard: messagesByShard,
	}

	res := testStruct.GetMessagesByShard()
	assert.Equal(t, messagesByShard, res)

	var nilStruct *GetReplicationMessagesResponse
	res = nilStruct.GetMessagesByShard()
	assert.Nil(t, res)
}

func TestCountDLQMessagesRequest_SerializeForLogging(t *testing.T) {
	// Test case where the struct is nil
	var nilStruct *CountDLQMessagesRequest
	res, err := nilStruct.SerializeForLogging()
	assert.Equal(t, "", res)
	assert.NoError(t, err)

	// Test case with a non-nil struct
	testStruct := CountDLQMessagesRequest{
		ForceFetch: true,
	}

	res, err = testStruct.SerializeForLogging()
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}

func TestCountDLQMessagesResponse(t *testing.T) {
	history := map[HistoryDLQCountKey]int64{
		{ShardID: 1, SourceCluster: "cluster-1"}: 100,
	}
	testStruct := CountDLQMessagesResponse{
		History: history,
		Domain:  200,
	}

	assert.Equal(t, history, testStruct.History)
	assert.Equal(t, int64(200), testStruct.Domain)

	// Test for empty history
	emptyStruct := CountDLQMessagesResponse{}
	assert.Nil(t, emptyStruct.History)
	assert.Equal(t, int64(0), emptyStruct.Domain)
}

func TestHistoryCountDLQMessagesResponse(t *testing.T) {
	entries := map[HistoryDLQCountKey]int64{
		{ShardID: 1, SourceCluster: "cluster-1"}: 100,
	}
	testStruct := HistoryCountDLQMessagesResponse{
		Entries: entries,
	}

	assert.Equal(t, entries, testStruct.Entries)

	// Test for empty entries
	emptyStruct := HistoryCountDLQMessagesResponse{}
	assert.Nil(t, emptyStruct.Entries)
}

func TestMergeDLQMessagesRequest_SerializeForLogging(t *testing.T) {
	// Test case where the struct is nil
	var nilStruct *MergeDLQMessagesRequest
	res, err := nilStruct.SerializeForLogging()
	assert.Equal(t, "", res)
	assert.NoError(t, err)

	// Test case with a non-nil struct
	dlqType := DLQTypeReplication
	endMessageID := int64(102)
	nextPageToken := []byte("token")
	testStruct := MergeDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "cluster-1",
		InclusiveEndMessageID: &endMessageID,
		MaximumPageSize:       50,
		NextPageToken:         nextPageToken,
	}

	res, err = testStruct.SerializeForLogging()
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}

func TestMergeDLQMessagesRequest_Getters(t *testing.T) {
	dlqType := DLQTypeReplication
	endMessageID := int64(102)
	nextPageToken := []byte("token")
	testStruct := MergeDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "cluster-1",
		InclusiveEndMessageID: &endMessageID,
		MaximumPageSize:       50,
		NextPageToken:         nextPageToken,
	}

	assert.Equal(t, dlqType, testStruct.GetType())
	assert.Equal(t, int32(101), testStruct.GetShardID())
	assert.Equal(t, "cluster-1", testStruct.GetSourceCluster())
	assert.Equal(t, endMessageID, testStruct.GetInclusiveEndMessageID())
	assert.Equal(t, int32(50), testStruct.GetMaximumPageSize())
	assert.Equal(t, nextPageToken, testStruct.GetNextPageToken())

	// Test for nil values
	var nilStruct *MergeDLQMessagesRequest
	assert.Equal(t, DLQType(0), nilStruct.GetType())
	assert.Equal(t, int32(0), nilStruct.GetShardID())
	assert.Equal(t, "", nilStruct.GetSourceCluster())
	assert.Equal(t, int64(0), nilStruct.GetInclusiveEndMessageID())
	assert.Equal(t, int32(0), nilStruct.GetMaximumPageSize())
	assert.Nil(t, nilStruct.GetNextPageToken())
}

func TestHistoryTaskV2Attributes_GetDomainID(t *testing.T) {
	testStruct := HistoryTaskV2Attributes{
		DomainID: "domain-id",
	}

	res := testStruct.GetDomainID()
	assert.Equal(t, "domain-id", res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetDomainID()
	assert.Equal(t, "", res)
}

func TestHistoryTaskV2Attributes_GetWorkflowID(t *testing.T) {
	testStruct := HistoryTaskV2Attributes{
		WorkflowID: "workflow-id",
	}

	res := testStruct.GetWorkflowID()
	assert.Equal(t, "workflow-id", res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetWorkflowID()
	assert.Equal(t, "", res)
}

func TestHistoryTaskV2Attributes_GetRunID(t *testing.T) {
	testStruct := HistoryTaskV2Attributes{
		RunID: "run-id",
	}

	res := testStruct.GetRunID()
	assert.Equal(t, "run-id", res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetRunID()
	assert.Equal(t, "", res)
}

func TestHistoryTaskV2Attributes_GetVersionHistoryItems(t *testing.T) {
	versionHistoryItems := []*VersionHistoryItem{{}, {}}
	testStruct := HistoryTaskV2Attributes{
		VersionHistoryItems: versionHistoryItems,
	}

	res := testStruct.GetVersionHistoryItems()
	assert.Equal(t, versionHistoryItems, res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetVersionHistoryItems()
	assert.Nil(t, res)
}

func TestHistoryTaskV2Attributes_GetEvents(t *testing.T) {
	events := &DataBlob{}
	testStruct := HistoryTaskV2Attributes{
		Events: events,
	}

	res := testStruct.GetEvents()
	assert.Equal(t, events, res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetEvents()
	assert.Nil(t, res)
}

func TestHistoryTaskV2Attributes_GetNewRunEvents(t *testing.T) {
	newRunEvents := &DataBlob{}
	testStruct := HistoryTaskV2Attributes{
		NewRunEvents: newRunEvents,
	}

	res := testStruct.GetNewRunEvents()
	assert.Equal(t, newRunEvents, res)

	var nilStruct *HistoryTaskV2Attributes
	res = nilStruct.GetNewRunEvents()
	assert.Nil(t, res)
}

func TestPurgeDLQMessagesRequest_SerializeForLogging(t *testing.T) {
	// Test case where the struct is nil
	var nilStruct *PurgeDLQMessagesRequest
	res, err := nilStruct.SerializeForLogging()
	assert.Equal(t, "", res)
	assert.NoError(t, err)

	// Test case with a non-nil struct
	dlqType := DLQTypeReplication
	endMessageID := int64(12345)
	testStruct := PurgeDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "test-cluster",
		InclusiveEndMessageID: &endMessageID,
	}

	res, err = testStruct.SerializeForLogging()
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}

func TestPurgeDLQMessagesRequest_Getters(t *testing.T) {
	dlqType := DLQTypeReplication
	endMessageID := int64(12345)
	testStruct := PurgeDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "test-cluster",
		InclusiveEndMessageID: &endMessageID,
	}

	assert.Equal(t, dlqType, testStruct.GetType())
	assert.Equal(t, int32(101), testStruct.GetShardID())
	assert.Equal(t, "test-cluster", testStruct.GetSourceCluster())
	assert.Equal(t, endMessageID, testStruct.GetInclusiveEndMessageID())

	// Test nil case
	var nilStruct *PurgeDLQMessagesRequest
	assert.Equal(t, DLQType(0), nilStruct.GetType())
	assert.Equal(t, int32(0), nilStruct.GetShardID())
	assert.Equal(t, "", nilStruct.GetSourceCluster())
	assert.Equal(t, int64(0), nilStruct.GetInclusiveEndMessageID())
}

func TestReadDLQMessagesRequest_SerializeForLogging(t *testing.T) {
	// Test case where the struct is nil
	var nilStruct *ReadDLQMessagesRequest
	res, err := nilStruct.SerializeForLogging()
	assert.Equal(t, "", res)
	assert.NoError(t, err)

	// Test case with a non-nil struct
	dlqType := DLQTypeReplication
	endMessageID := int64(12345)
	nextPageToken := []byte("token")
	testStruct := ReadDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "test-cluster",
		InclusiveEndMessageID: &endMessageID,
		MaximumPageSize:       50,
		NextPageToken:         nextPageToken,
	}

	res, err = testStruct.SerializeForLogging()
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}

func TestReadDLQMessagesRequest_Getters(t *testing.T) {
	dlqType := DLQTypeReplication
	endMessageID := int64(12345)
	nextPageToken := []byte("token")
	testStruct := ReadDLQMessagesRequest{
		Type:                  &dlqType,
		ShardID:               101,
		SourceCluster:         "test-cluster",
		InclusiveEndMessageID: &endMessageID,
		MaximumPageSize:       50,
		NextPageToken:         nextPageToken,
	}

	assert.Equal(t, dlqType, testStruct.GetType())
	assert.Equal(t, int32(101), testStruct.GetShardID())
	assert.Equal(t, "test-cluster", testStruct.GetSourceCluster())
	assert.Equal(t, endMessageID, testStruct.GetInclusiveEndMessageID())
	assert.Equal(t, int32(50), testStruct.GetMaximumPageSize())
	assert.Equal(t, nextPageToken, testStruct.GetNextPageToken())

	// Test nil case
	var nilStruct *ReadDLQMessagesRequest
	assert.Equal(t, DLQType(0), nilStruct.GetType())
	assert.Equal(t, int32(0), nilStruct.GetShardID())
	assert.Equal(t, "", nilStruct.GetSourceCluster())
	assert.Equal(t, int64(0), nilStruct.GetInclusiveEndMessageID())
	assert.Equal(t, int32(0), nilStruct.GetMaximumPageSize())
	assert.Nil(t, nilStruct.GetNextPageToken())
}

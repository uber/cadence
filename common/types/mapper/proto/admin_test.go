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

package proto

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	adminv1 "github.com/uber/cadence-idl/go/proto/admin/v1"
	v1 "github.com/uber/cadence-idl/go/proto/api/v1"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
)

func TestAdminAddSearchAttributeRequest(t *testing.T) {
	for _, item := range []*types.AddSearchAttributeRequest{nil, {}, &testdata.AdminAddSearchAttributeRequest} {
		assert.Equal(t, item, ToAdminAddSearchAttributeRequest(FromAdminAddSearchAttributeRequest(item)))
	}
}
func TestAdminCloseShardRequest(t *testing.T) {
	for _, item := range []*types.CloseShardRequest{nil, {}, &testdata.AdminCloseShardRequest} {
		assert.Equal(t, item, ToAdminCloseShardRequest(FromAdminCloseShardRequest(item)))
	}
}
func TestAdminDescribeClusterResponse(t *testing.T) {
	for _, item := range []*types.DescribeClusterResponse{nil, {}, &testdata.AdminDescribeClusterResponse} {
		assert.Equal(t, item, ToAdminDescribeClusterResponse(FromAdminDescribeClusterResponse(item)))
	}
}
func TestAdminDescribeHistoryHostRequest(t *testing.T) {
	for _, item := range []*types.DescribeHistoryHostRequest{
		nil,
		&testdata.AdminDescribeHistoryHostRequest_ByHost,
		&testdata.AdminDescribeHistoryHostRequest_ByShard,
		&testdata.AdminDescribeHistoryHostRequest_ByExecution,
	} {
		assert.Equal(t, item, ToAdminDescribeHistoryHostRequest(FromAdminDescribeHistoryHostRequest(item)))
	}
	assert.Panics(t, func() { ToAdminDescribeHistoryHostRequest(&adminv1.DescribeHistoryHostRequest{}) })
	assert.Panics(t, func() { FromAdminDescribeHistoryHostRequest(&types.DescribeHistoryHostRequest{}) })
}
func TestAdminDescribeHistoryHostResponse(t *testing.T) {
	for _, item := range []*types.DescribeHistoryHostResponse{nil, {}, &testdata.AdminDescribeHistoryHostResponse} {
		assert.Equal(t, item, ToAdminDescribeHistoryHostResponse(FromAdminDescribeHistoryHostResponse(item)))
	}
}
func TestAdminDescribeQueueRequest(t *testing.T) {
	for _, item := range []*types.DescribeQueueRequest{nil, {}, &testdata.AdminDescribeQueueRequest} {
		assert.Equal(t, item, ToAdminDescribeQueueRequest(FromAdminDescribeQueueRequest(item)))
	}
}
func TestAdminDescribeQueueResponse(t *testing.T) {
	for _, item := range []*types.DescribeQueueResponse{nil, {}, &testdata.AdminDescribeQueueResponse} {
		assert.Equal(t, item, ToAdminDescribeQueueResponse(FromAdminDescribeQueueResponse(item)))
	}
}
func TestAdminDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*types.AdminDescribeWorkflowExecutionRequest{nil, {}, &testdata.AdminDescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, ToAdminDescribeWorkflowExecutionRequest(FromAdminDescribeWorkflowExecutionRequest(item)))
	}
}
func TestAdminDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*types.AdminDescribeWorkflowExecutionResponse{nil, {ShardID: "0"}, &testdata.AdminDescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, ToAdminDescribeWorkflowExecutionResponse(FromAdminDescribeWorkflowExecutionResponse(item)))
	}
}
func TestAdminGetDLQReplicationMessagesRequest(t *testing.T) {
	for _, item := range []*types.GetDLQReplicationMessagesRequest{nil, {}, &testdata.AdminGetDLQReplicationMessagesRequest} {
		assert.Equal(t, item, ToAdminGetDLQReplicationMessagesRequest(FromAdminGetDLQReplicationMessagesRequest(item)))
	}
}
func TestAdminGetDLQReplicationMessagesResponse(t *testing.T) {
	for _, item := range []*types.GetDLQReplicationMessagesResponse{nil, {}, &testdata.AdminGetDLQReplicationMessagesResponse} {
		assert.Equal(t, item, ToAdminGetDLQReplicationMessagesResponse(FromAdminGetDLQReplicationMessagesResponse(item)))
	}
}
func TestAdminGetDomainReplicationMessagesRequest(t *testing.T) {
	for _, item := range []*types.GetDomainReplicationMessagesRequest{nil, {}, &testdata.AdminGetDomainReplicationMessagesRequest} {
		assert.Equal(t, item, ToAdminGetDomainReplicationMessagesRequest(FromAdminGetDomainReplicationMessagesRequest(item)))
	}
}
func TestAdminGetDomainReplicationMessagesResponse(t *testing.T) {
	for _, item := range []*types.GetDomainReplicationMessagesResponse{nil, {}, &testdata.AdminGetDomainReplicationMessagesResponse} {
		assert.Equal(t, item, ToAdminGetDomainReplicationMessagesResponse(FromAdminGetDomainReplicationMessagesResponse(item)))
	}
}
func TestAdminGetReplicationMessagesRequest(t *testing.T) {
	for _, item := range []*types.GetReplicationMessagesRequest{nil, {}, &testdata.AdminGetReplicationMessagesRequest} {
		assert.Equal(t, item, ToAdminGetReplicationMessagesRequest(FromAdminGetReplicationMessagesRequest(item)))
	}
}
func TestAdminGetReplicationMessagesResponse(t *testing.T) {
	for _, item := range []*types.GetReplicationMessagesResponse{nil, {}, &testdata.AdminGetReplicationMessagesResponse} {
		assert.Equal(t, item, ToAdminGetReplicationMessagesResponse(FromAdminGetReplicationMessagesResponse(item)))
	}
}
func TestAdminGetWorkflowExecutionRawHistoryV2Request(t *testing.T) {
	for _, item := range []*types.GetWorkflowExecutionRawHistoryV2Request{nil, {}, &testdata.AdminGetWorkflowExecutionRawHistoryV2Request} {
		assert.Equal(t, item, ToAdminGetWorkflowExecutionRawHistoryV2Request(FromAdminGetWorkflowExecutionRawHistoryV2Request(item)))
	}
}
func TestAdminGetWorkflowExecutionRawHistoryV2Response(t *testing.T) {
	for _, item := range []*types.GetWorkflowExecutionRawHistoryV2Response{nil, {}, &testdata.AdminGetWorkflowExecutionRawHistoryV2Response} {
		assert.Equal(t, item, ToAdminGetWorkflowExecutionRawHistoryV2Response(FromAdminGetWorkflowExecutionRawHistoryV2Response(item)))
	}
}
func TestAdminCountDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.CountDLQMessagesRequest{nil, {}, &testdata.AdminCountDLQMessagesRequest} {
		assert.Equal(t, item, ToAdminCountDLQMessagesRequest(FromAdminCountDLQMessagesRequest(item)))
	}
}
func TestAdminCountDLQMessagesResponse(t *testing.T) {
	for _, item := range []*types.CountDLQMessagesResponse{nil, {}, &testdata.AdminCountDLQMessagesResponse} {
		assert.Equal(t, item, ToAdminCountDLQMessagesResponse(FromAdminCountDLQMessagesResponse(item)))
	}
}
func TestAdminMergeDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.MergeDLQMessagesRequest{nil, {}, &testdata.AdminMergeDLQMessagesRequest} {
		assert.Equal(t, item, ToAdminMergeDLQMessagesRequest(FromAdminMergeDLQMessagesRequest(item)))
	}
}
func TestAdminMergeDLQMessagesResponse(t *testing.T) {
	for _, item := range []*types.MergeDLQMessagesResponse{nil, {}, &testdata.AdminMergeDLQMessagesResponse} {
		assert.Equal(t, item, ToAdminMergeDLQMessagesResponse(FromAdminMergeDLQMessagesResponse(item)))
	}
}
func TestAdminPurgeDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.PurgeDLQMessagesRequest{nil, {}, &testdata.AdminPurgeDLQMessagesRequest} {
		assert.Equal(t, item, ToAdminPurgeDLQMessagesRequest(FromAdminPurgeDLQMessagesRequest(item)))
	}
}
func TestAdminReadDLQMessagesRequest(t *testing.T) {
	for _, item := range []*types.ReadDLQMessagesRequest{nil, {}, &testdata.AdminReadDLQMessagesRequest} {
		assert.Equal(t, item, ToAdminReadDLQMessagesRequest(FromAdminReadDLQMessagesRequest(item)))
	}
}
func TestAdminReadDLQMessagesResponse(t *testing.T) {
	for _, item := range []*types.ReadDLQMessagesResponse{nil, {}, &testdata.AdminReadDLQMessagesResponse} {
		assert.Equal(t, item, ToAdminReadDLQMessagesResponse(FromAdminReadDLQMessagesResponse(item)))
	}
}
func TestAdminReapplyEventsRequest(t *testing.T) {
	for _, item := range []*types.ReapplyEventsRequest{nil, {}, &testdata.AdminReapplyEventsRequest} {
		assert.Equal(t, item, ToAdminReapplyEventsRequest(FromAdminReapplyEventsRequest(item)))
	}
}
func TestAdminRefreshWorkflowTasksRequest(t *testing.T) {
	for _, item := range []*types.RefreshWorkflowTasksRequest{nil, {}, &testdata.AdminRefreshWorkflowTasksRequest} {
		assert.Equal(t, item, ToAdminRefreshWorkflowTasksRequest(FromAdminRefreshWorkflowTasksRequest(item)))
	}
}
func TestAdminRemoveTaskRequest(t *testing.T) {
	for _, item := range []*types.RemoveTaskRequest{nil, {}, &testdata.AdminRemoveTaskRequest} {
		assert.Equal(t, item, ToAdminRemoveTaskRequest(FromAdminRemoveTaskRequest(item)))
	}
}
func TestAdminResendReplicationTasksRequest(t *testing.T) {
	for _, item := range []*types.ResendReplicationTasksRequest{nil, {}, &testdata.AdminResendReplicationTasksRequest} {
		assert.Equal(t, item, ToAdminResendReplicationTasksRequest(FromAdminResendReplicationTasksRequest(item)))
	}
}
func TestAdminResetQueueRequest(t *testing.T) {
	for _, item := range []*types.ResetQueueRequest{nil, {}, &testdata.AdminResetQueueRequest} {
		assert.Equal(t, item, ToAdminResetQueueRequest(FromAdminResetQueueRequest(item)))
	}
}

func TestAdminGetCrossClusterTasksRequest(t *testing.T) {
	for _, item := range []*types.GetCrossClusterTasksRequest{nil, {}, &testdata.AdminGetCrossClusterTasksRequest} {
		assert.Equal(t, item, ToAdminGetCrossClusterTasksRequest(FromAdminGetCrossClusterTasksRequest(item)))
	}
}

func TestAdminGetCrossClusterTasksResponse(t *testing.T) {
	for _, item := range []*types.GetCrossClusterTasksResponse{nil, {}, &testdata.AdminGetCrossClusterTasksResponse} {
		assert.Equal(t, item, ToAdminGetCrossClusterTasksResponse(FromAdminGetCrossClusterTasksResponse(item)))
	}
}

func TestAdminRespondCrossClusterTasksCompletedRequest(t *testing.T) {
	for _, item := range []*types.RespondCrossClusterTasksCompletedRequest{nil, {}, &testdata.AdminRespondCrossClusterTasksCompletedRequest} {
		assert.Equal(t, item, ToAdminRespondCrossClusterTasksCompletedRequest(FromAdminRespondCrossClusterTasksCompletedRequest(item)))
	}
}

func TestAdminRespondCrossClusterTasksCompletedResponse(t *testing.T) {
	for _, item := range []*types.RespondCrossClusterTasksCompletedResponse{nil, {}, &testdata.AdminRespondCrossClusterTasksCompletedResponse} {
		assert.Equal(t, item, ToAdminRespondCrossClusterTasksCompletedResponse(FromAdminRespondCrossClusterTasksCompletedResponse(item)))
	}
}

func TestFromGetGlobalIsolationGroupsResponse(t *testing.T) {
	tests := map[string]struct {
		in       *types.GetGlobalIsolationGroupsResponse
		expected *adminv1.GetGlobalIsolationGroupsResponse
	}{
		"Valid mapping": {
			in: &types.GetGlobalIsolationGroupsResponse{
				IsolationGroups: types.IsolationGroupConfiguration{
					"zone 0": {
						Name:  "zone 0",
						State: types.IsolationGroupStateHealthy,
					},
					"zone 1": {
						Name:  "zone 1",
						State: types.IsolationGroupStateDrained,
					},
				},
			},
			expected: &adminv1.GetGlobalIsolationGroupsResponse{
				IsolationGroups: &v1.IsolationGroupConfiguration{
					IsolationGroups: []*v1.IsolationGroupPartition{
						{
							Name:  "zone 0",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_HEALTHY,
						},
						{
							Name:  "zone 1",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_DRAINED,
						},
					},
				},
			},
		},
		"nil - 1": {
			in: &types.GetGlobalIsolationGroupsResponse{
				IsolationGroups: types.IsolationGroupConfiguration{},
			},
			expected: &adminv1.GetGlobalIsolationGroupsResponse{
				IsolationGroups: &v1.IsolationGroupConfiguration{},
			},
		},
		"nil - 2": {
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res := FromGetGlobalIsolationGroupsResponse(td.in)
			assert.Equal(t, td.expected, res, "mapping")
			roundTrip := ToGetGlobalIsolationGroupsResponse(res)
			if td.in != nil {
				assert.Equal(t, td.in, roundTrip, "roundtrip")
			}
		})
	}
}

func TestToGetGlobalIsolationGroupsRequest(t *testing.T) {

	tests := map[string]struct {
		in       *adminv1.GetGlobalIsolationGroupsRequest
		expected *types.GetGlobalIsolationGroupsRequest
	}{
		"Valid mapping": {
			in:       &adminv1.GetGlobalIsolationGroupsRequest{},
			expected: &types.GetGlobalIsolationGroupsRequest{},
		},
		"nil - 2": {
			in:       &adminv1.GetGlobalIsolationGroupsRequest{},
			expected: &types.GetGlobalIsolationGroupsRequest{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, ToGetGlobalIsolationGroupsRequest(td.in))
		})
	}
}

func TestFromGetDomainIsolationGroupsResponse(t *testing.T) {
	tests := map[string]struct {
		in       *types.GetDomainIsolationGroupsResponse
		expected *adminv1.GetDomainIsolationGroupsResponse
	}{
		"Valid mapping": {
			in: &types.GetDomainIsolationGroupsResponse{
				IsolationGroups: types.IsolationGroupConfiguration{
					"zone 0": {
						Name:  "zone 0",
						State: types.IsolationGroupStateHealthy,
					},
					"zone 1": {
						Name:  "zone 1",
						State: types.IsolationGroupStateDrained,
					},
				},
			},
			expected: &adminv1.GetDomainIsolationGroupsResponse{
				IsolationGroups: &v1.IsolationGroupConfiguration{
					IsolationGroups: []*v1.IsolationGroupPartition{
						{
							Name:  "zone 0",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_HEALTHY,
						},
						{
							Name:  "zone 1",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_DRAINED,
						},
					},
				},
			},
		},
		"empty": {
			in: &types.GetDomainIsolationGroupsResponse{
				IsolationGroups: types.IsolationGroupConfiguration{},
			},
			expected: &adminv1.GetDomainIsolationGroupsResponse{
				IsolationGroups: &v1.IsolationGroupConfiguration{
					IsolationGroups: []*v1.IsolationGroupPartition{},
				},
			},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res := FromGetDomainIsolationGroupsResponse(td.in)
			// map iteration is nondeterministic
			sort.Slice(res.IsolationGroups.IsolationGroups, func(i int, j int) bool {
				return res.IsolationGroups.IsolationGroups[i].Name > res.IsolationGroups.IsolationGroups[j].Name
			})
		})
	}
}

func TestToGetDomainIsolationGroupsRequest(t *testing.T) {

	tests := map[string]struct {
		in       *adminv1.GetDomainIsolationGroupsRequest
		expected *types.GetDomainIsolationGroupsRequest
	}{
		"Valid mapping": {
			in: &adminv1.GetDomainIsolationGroupsRequest{
				Domain: "domain123",
			},
			expected: &types.GetDomainIsolationGroupsRequest{
				Domain: "domain123",
			},
		},
		"empty": {
			in:       &adminv1.GetDomainIsolationGroupsRequest{},
			expected: &types.GetDomainIsolationGroupsRequest{},
		},
		"nil": {
			in:       nil,
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, ToGetDomainIsolationGroupsRequest(td.in))
		})
	}
}

func TestFromUpdateGlobalIsolationGroupsResponse(t *testing.T) {

	tests := map[string]struct {
		in       *types.UpdateGlobalIsolationGroupsResponse
		expected *adminv1.UpdateGlobalIsolationGroupsResponse
	}{
		"Valid mapping": {},
		"empty": {
			in:       &types.UpdateGlobalIsolationGroupsResponse{},
			expected: &adminv1.UpdateGlobalIsolationGroupsResponse{},
		},
		"nil": {
			in:       nil,
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, FromUpdateGlobalIsolationGroupsResponse(td.in))
		})
	}
}

func TestToUpdateGlobalIsolationGroupsRequest(t *testing.T) {

	tests := map[string]struct {
		in       *adminv1.UpdateGlobalIsolationGroupsRequest
		expected *types.UpdateGlobalIsolationGroupsRequest
	}{
		"Valid mapping": {
			in: &adminv1.UpdateGlobalIsolationGroupsRequest{
				IsolationGroups: &v1.IsolationGroupConfiguration{
					IsolationGroups: []*v1.IsolationGroupPartition{
						{
							Name:  "zone 1",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_HEALTHY,
						},
						{
							Name:  "zone 2",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_DRAINED,
						},
					},
				},
			},
			expected: &types.UpdateGlobalIsolationGroupsRequest{
				IsolationGroups: types.IsolationGroupConfiguration{
					"zone 1": types.IsolationGroupPartition{
						Name:  "zone 1",
						State: types.IsolationGroupStateHealthy,
					},
					"zone 2": types.IsolationGroupPartition{
						Name:  "zone 2",
						State: types.IsolationGroupStateDrained,
					},
				},
			},
		},
		"empty": {
			in:       &adminv1.UpdateGlobalIsolationGroupsRequest{},
			expected: &types.UpdateGlobalIsolationGroupsRequest{},
		},
		"nil": {
			in:       nil,
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res := ToUpdateGlobalIsolationGroupsRequest(td.in)
			assert.Equal(t, td.expected, res, "conversion")
			roundTrip := FromUpdateGlobalIsolationGroupsRequest(res)
			if td.in != nil {
				assert.Equal(t, td.in, roundTrip, "roundtrip")
			}
		})
	}
}

func TestFromUpdateDomainIsolationGroupsResponse(t *testing.T) {

	tests := map[string]struct {
		in       *types.UpdateDomainIsolationGroupsResponse
		expected *adminv1.UpdateDomainIsolationGroupsResponse
	}{
		"empty": {
			in:       &types.UpdateDomainIsolationGroupsResponse{},
			expected: &adminv1.UpdateDomainIsolationGroupsResponse{},
		},
		"nil": {
			in:       nil,
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, FromUpdateDomainIsolationGroupsResponse(td.in))
		})
	}
}

func TestToUpdateDomainIsolationGroupsRequest(t *testing.T) {

	tests := map[string]struct {
		in       *adminv1.UpdateDomainIsolationGroupsRequest
		expected *types.UpdateDomainIsolationGroupsRequest
	}{
		"valid": {
			in: &adminv1.UpdateDomainIsolationGroupsRequest{
				Domain: "test-domain",
				IsolationGroups: &v1.IsolationGroupConfiguration{
					IsolationGroups: []*v1.IsolationGroupPartition{
						{
							Name:  "zone-1",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_HEALTHY,
						},
						{
							Name:  "zone-2",
							State: v1.IsolationGroupState_ISOLATION_GROUP_STATE_DRAINED,
						},
					},
				},
			},
			expected: &types.UpdateDomainIsolationGroupsRequest{
				Domain: "test-domain",
				IsolationGroups: types.IsolationGroupConfiguration{
					"zone-1": types.IsolationGroupPartition{
						Name:  "zone-1",
						State: types.IsolationGroupStateHealthy,
					},
					"zone-2": types.IsolationGroupPartition{
						Name:  "zone-2",
						State: types.IsolationGroupStateDrained,
					},
				},
			},
		},
		"nil": {
			in:       nil,
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, ToUpdateDomainIsolationGroupsRequest(td.in))
		})
	}
}

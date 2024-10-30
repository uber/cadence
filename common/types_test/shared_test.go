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

package types_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/types"
)

func TestSerializeRequestForLogging(t *testing.T) {
	requestTypes := []authorization.FilteredRequestBody{
		// Admin requests
		&types.AddSearchAttributeRequest{},
		&types.AdminDescribeWorkflowExecutionRequest{},
		&types.GetWorkflowExecutionRawHistoryV2Request{},
		&types.ResendReplicationTasksRequest{},
		&types.GetDynamicConfigRequest{},
		&types.UpdateDynamicConfigRequest{},
		&types.RestoreDynamicConfigRequest{},
		&types.AdminDeleteWorkflowRequest{},
		&types.ListDynamicConfigRequest{},
		&types.GetGlobalIsolationGroupsRequest{},
		&types.UpdateGlobalIsolationGroupsRequest{},
		&types.GetDomainIsolationGroupsRequest{},
		&types.UpdateDomainIsolationGroupsRequest{},
		&types.GetDomainAsyncWorkflowConfiguratonRequest{},
		&types.UpdateDomainAsyncWorkflowConfiguratonRequest{},

		// General request types
		&types.CloseShardRequest{},
		&types.CountWorkflowExecutionsRequest{},
		&types.DeprecateDomainRequest{},
		&types.DescribeDomainRequest{},
		&types.DescribeHistoryHostRequest{},
		&types.DescribeShardDistributionRequest{},
		&types.DescribeQueueRequest{},
		&types.DescribeTaskListRequest{},
		&types.DescribeWorkflowExecutionRequest{},
		&types.GetWorkflowExecutionHistoryRequest{},
		&types.ListArchivedWorkflowExecutionsRequest{},
		&types.ListClosedWorkflowExecutionsRequest{},
		&types.ListDomainsRequest{},
		&types.ListOpenWorkflowExecutionsRequest{},
		&types.ListTaskListPartitionsRequest{},
		&types.GetTaskListsByDomainRequest{},
		&types.ListWorkflowExecutionsRequest{},
		&types.PollForActivityTaskRequest{},
		&types.PollForDecisionTaskRequest{},
		&types.QueryWorkflowRequest{},
		&types.ReapplyEventsRequest{},
		&types.RecordActivityTaskHeartbeatRequest{},
		&types.RefreshWorkflowTasksRequest{},
		&types.RegisterDomainRequest{},
		&types.RemoveTaskRequest{},
		&types.RequestCancelWorkflowExecutionRequest{},
		&types.ResetQueueRequest{},
		&types.ResetStickyTaskListRequest{},
		&types.ResetWorkflowExecutionRequest{},
		&types.SignalWithStartWorkflowExecutionRequest{},
		&types.SignalWorkflowExecutionRequest{},
		&types.RestartWorkflowExecutionRequest{},
		&types.DiagnoseWorkflowExecutionRequest{},
		&types.StartWorkflowExecutionRequest{},
		&types.TerminateWorkflowExecutionRequest{},
		&types.UpdateDomainRequest{},
		&types.GetCrossClusterTasksRequest{},
		&types.RespondCrossClusterTasksCompletedRequest{},

		// Replicator
		&types.GetDLQReplicationMessagesRequest{},
		&types.GetDomainReplicationMessagesRequest{},
		&types.GetReplicationMessagesRequest{},
		&types.CountDLQMessagesRequest{},
		&types.MergeDLQMessagesRequest{},
		&types.PurgeDLQMessagesRequest{},
		&types.ReadDLQMessagesRequest{},
	}

	for _, requestType := range requestTypes {
		typ := reflect.TypeOf(requestType)
		name := typ.Elem().Name()
		t.Run(name, func(t *testing.T) {
			// Test that the call with a real request returns as expected
			serializedBody, err := requestType.SerializeForLogging()
			assert.NoError(t, err)

			assert.NotEmpty(t, serializedBody)

			// Test that nil does not error, but returns empty string
			nilPointer := reflect.Zero(typ).Interface()
			serializedBody, err = nilPointer.(authorization.FilteredRequestBody).SerializeForLogging()
			assert.NoError(t, err)
			assert.Empty(t, serializedBody)
		})
	}
}

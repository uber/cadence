// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

func TestValidateDomainUUID(t *testing.T) {
	testCases := []struct {
		msg        string
		domainUUID string
		valid      bool
	}{
		{
			msg:        "empty",
			domainUUID: "",
			valid:      false,
		},
		{
			msg:        "invalid",
			domainUUID: "some random uuid",
			valid:      false,
		},
		{
			msg:        "valid",
			domainUUID: uuid.New(),
			valid:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := ValidateDomainUUID(tc.domainUUID)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetActiveDomainEntry(t *testing.T) {
	testCases := []struct {
		msg              string
		domainCacheEntry *cache.DomainCacheEntry
		expectErr        bool
	}{
		{
			msg:              "local domain",
			domainCacheEntry: constants.TestLocalDomainEntry,
			expectErr:        false,
		},
		{
			msg: "active global domain",
			domainCacheEntry: cache.NewGlobalDomainCacheEntryForTest(
				&persistence.DomainInfo{ID: constants.TestDomainID, Name: constants.TestDomainName},
				nil,
				&persistence.DomainReplicationConfig{
					ActiveClusterName: cluster.TestCurrentClusterName,
					Clusters: []*persistence.ClusterReplicationConfig{
						{ClusterName: cluster.TestCurrentClusterName},
						{ClusterName: cluster.TestAlternativeClusterName},
					},
				},
				constants.TestVersion,
				cluster.GetTestClusterMetadata(true, true),
			),
			expectErr: false,
		},
		{
			msg: "passive global domain",
			domainCacheEntry: cache.NewGlobalDomainCacheEntryForTest(
				&persistence.DomainInfo{ID: constants.TestDomainID, Name: constants.TestDomainName},
				nil,
				&persistence.DomainReplicationConfig{
					ActiveClusterName: cluster.TestCurrentClusterName,
					Clusters: []*persistence.ClusterReplicationConfig{
						{ClusterName: cluster.TestCurrentClusterName},
						{ClusterName: cluster.TestAlternativeClusterName},
					},
				},
				constants.TestVersion,
				cluster.NewMetadata(
					loggerimpl.NewNopLogger(),
					dynamicconfig.GetBoolPropertyFn(true),
					int64(10),
					cluster.TestCurrentClusterName,
					cluster.TestAlternativeClusterName,
					cluster.TestAllClusterInfo,
				),
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			controller := gomock.NewController(t)
			mockShard := shard.NewTestContext(
				controller,
				&persistence.ShardInfo{
					ShardID: 10,
					RangeID: 1,
				},
				nil,
			)
			mockDomainCache := mockShard.Resource.DomainCache

			domainID := tc.domainCacheEntry.GetInfo().ID
			mockDomainCache.EXPECT().GetDomainByID(domainID).Return(tc.domainCacheEntry, nil).Times(1)

			entry, err := GetActiveDomainEntry(mockShard, domainID)
			if tc.expectErr {
				require.Error(t, err)
				require.IsType(t, &types.DomainNotActiveError{}, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.domainCacheEntry, entry)

			controller.Finish()
		})
	}

}

func TestUpdateHelper(t *testing.T) {
	testCases := []struct {
		msg         string
		mockSetupFn func(*execution.MockContext, *execution.MockMutableState)
		actionFn    UpdateActionFunc
	}{
		{
			msg: "stale mutable state",
			mockSetupFn: func(mockContext *execution.MockContext, mockMutableState *execution.MockMutableState) {
				mockContext.EXPECT().Clear().Times(1)
				mockMutableState.EXPECT().GetNextEventID().Return(common.FirstEventID).Times(1)
				mockMutableState.EXPECT().GetNextEventID().Return(common.FirstEventID + 1).Times(1)
			},
			actionFn: func(context execution.Context, mutableState execution.MutableState) (*UpdateAction, error) {
				if mutableState.GetNextEventID() == common.FirstEventID {
					return nil, ErrStaleState
				}
				return &UpdateAction{Noop: true}, nil
			},
		},
		{
			msg: "schedule new decision",
			mockSetupFn: func(mockContext *execution.MockContext, mockMutableState *execution.MockMutableState) {
				mockMutableState.EXPECT().HasPendingDecision().Return(false).Times(1)
				mockMutableState.EXPECT().AddDecisionTaskScheduledEvent(gomock.Any()).Return(&execution.DecisionInfo{}, nil).Times(1)
				mockContext.EXPECT().UpdateWorkflowExecutionAsActive(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			actionFn: func(context execution.Context, mutableState execution.MutableState) (*UpdateAction, error) {
				return UpdateWithNewDecision, nil
			},
		},
		{
			msg: "update workflow conflict",
			mockSetupFn: func(mockContext *execution.MockContext, mockMutableState *execution.MockMutableState) {
				mockContext.EXPECT().UpdateWorkflowExecutionAsActive(gomock.Any(), gomock.Any()).Return(execution.ErrConflict).Times(ConditionalRetryCount - 1)
				mockContext.EXPECT().UpdateWorkflowExecutionAsActive(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			actionFn: func(context execution.Context, mutableState execution.MutableState) (*UpdateAction, error) {
				return UpdateWithoutDecision, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			controller := gomock.NewController(t)
			mockMutableState := execution.NewMockMutableState(controller)
			mockContext := execution.NewMockContext(controller)
			mockContext.EXPECT().LoadWorkflowExecution(gomock.Any()).Return(mockMutableState, nil).AnyTimes()
			workflowContext := NewContext(mockContext, nil, mockMutableState)

			tc.mockSetupFn(mockContext, mockMutableState)
			err := updateHelper(context.Background(), workflowContext, time.Now(), tc.actionFn)
			require.NoError(t, err)

			controller.Finish()
		})
	}

}

func TestWorkflowLoad(t *testing.T) {
	persistenceMS := &persistence.WorkflowMutableState{
		ExecutionInfo: &persistence.WorkflowExecutionInfo{
			DomainID:   constants.TestDomainID,
			WorkflowID: constants.TestWorkflowID,
			RunID:      constants.TestRunID,
			State:      persistence.WorkflowStateRunning,
		},
		ExecutionStats: &persistence.ExecutionStats{},
	}

	testCases := []struct {
		msg         string
		runID       string
		mockSetupFn func(*shard.TestContext)
	}{
		{
			msg:   "runID not empty",
			runID: constants.TestRunID,
			mockSetupFn: func(mockShard *shard.TestContext) {
				mockShard.Resource.ExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(
					&persistence.GetWorkflowExecutionResponse{
						State: persistenceMS,
					},
					nil,
				).Times(1)
			},
		},
		{
			msg:   "current run closed",
			runID: "",
			mockSetupFn: func(mockShard *shard.TestContext) {
				persistenceMS.ExecutionInfo.State = persistence.WorkflowStateCompleted
				mockShard.Resource.ExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(
					&persistence.GetWorkflowExecutionResponse{
						State: persistenceMS,
					},
					nil,
				).Times(1)
				mockShard.Resource.ExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(
					&persistence.GetCurrentExecutionResponse{
						RunID: constants.TestRunID,
					},
					nil,
				).Times(2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			controller := gomock.NewController(t)
			mockShard := shard.NewTestContext(
				controller,
				&persistence.ShardInfo{
					ShardID: 10,
					RangeID: 1,
				},
				config.NewForTest(),
			)

			mockDomainCache := mockShard.Resource.DomainCache
			mockDomainCache.EXPECT().GetDomainByID(constants.TestLocalDomainEntry.GetInfo().ID).Return(constants.TestLocalDomainEntry, nil)
			mockDomainCache.EXPECT().GetDomainName(constants.TestLocalDomainEntry.GetInfo().ID).Return(constants.TestLocalDomainEntry.GetInfo().Name, nil)

			tc.mockSetupFn(mockShard)

			workflowContext, err := Load(
				context.Background(),
				execution.NewCache(mockShard),
				mockShard.Resource.ExecutionMgr,
				constants.TestDomainID,
				constants.TestWorkflowID,
				tc.runID,
			)
			require.NoError(t, err)
			require.Equal(t, constants.TestWorkflowID, workflowContext.GetWorkflowID())
			require.Equal(t, constants.TestRunID, workflowContext.GetRunID())
			workflowContext.GetReleaseFn()(nil)

			controller.Finish()
		})
	}
}

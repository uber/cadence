// Copyright (c) 2024 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

func TestGetResurrectedTimers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mockShard *shard.MockContext, mockMutableState *MockMutableState, domainCache *cache.MockDomainCache, manager *persistence.MockHistoryManager)
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "No pending timers",
			setup: func(mockShard *shard.MockContext, mockMutableState *MockMutableState, domainCache *cache.MockDomainCache, manager *persistence.MockHistoryManager) {
				mockMutableState.EXPECT().GetPendingTimerInfos().Return(map[string]*persistence.TimerInfo{}).Times(1)
			},
			want: map[string]struct{}{},
		},
		{
			name: "Timers with no corresponding events",
			setup: func(mockShard *shard.MockContext, mockMutableState *MockMutableState, domainCache *cache.MockDomainCache, manager *persistence.MockHistoryManager) {
				mockMutableState.EXPECT().GetPendingTimerInfos().Return(map[string]*persistence.TimerInfo{
					"timer1": {
						TimerID:    "timer1",
						ExpiryTime: clock.NewRealTimeSource().Now().Add(10 * time.Minute),
					},
				}).Times(1)
				mockMutableState.EXPECT().GetCurrentBranchToken().Return([]byte("branchToken"), nil).Times(1)
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{DomainID: "testDomain"}).Times(1)
				mockMutableState.EXPECT().GetNextEventID().Return(int64(10)).Times(1)

				mockShard.EXPECT().GetHistoryManager().Return(manager).Times(1)
				manager.EXPECT().GetHistoryTree(gomock.Any(), gomock.Any()).Return(&persistence.GetHistoryTreeResponse{
					Branches: []*workflow.HistoryBranch{
						{TreeID: common.StringPtr("treeID1"), BranchID: common.StringPtr("branchID1"), Ancestors: []*workflow.HistoryBranchRange{}},
					},
				}, nil).AnyTimes()

				manager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{
					HistoryEvents: []*types.HistoryEvent{},
				}, nil).Times(1)

				mockShard.EXPECT().GetShardID().Return(1).Times(1)
				mockShard.EXPECT().GetDomainCache().Return(domainCache).Times(1)
				domainCache.EXPECT().GetDomainName("testDomain").Return("Test Domain", nil).Times(1)
			},
			want: map[string]struct{}{},
		},
		{
			name: "Error on fetching branch token",
			setup: func(mockShard *shard.MockContext, mockMutableState *MockMutableState, domainCache *cache.MockDomainCache, manager *persistence.MockHistoryManager) {
				mockMutableState.EXPECT().GetPendingTimerInfos().Return(map[string]*persistence.TimerInfo{"timer1": {
					TimerID:    "timer1",
					ExpiryTime: clock.NewRealTimeSource().Now().Add(10 * time.Minute),
				}}).Times(1)
				mockMutableState.EXPECT().GetCurrentBranchToken().Return(nil, errors.New("error fetching token")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Processing multiple events",
			setup: func(mockShard *shard.MockContext, mockMutableState *MockMutableState, domainCache *cache.MockDomainCache, manager *persistence.MockHistoryManager) {
				mockMutableState.EXPECT().GetPendingTimerInfos().Return(map[string]*persistence.TimerInfo{
					"timer1": {TimerID: "timer1", ExpiryTime: clock.NewRealTimeSource().Now().Add(10 * time.Minute)},
					"timer2": {TimerID: "timer2", ExpiryTime: clock.NewRealTimeSource().Now().Add(10 * time.Minute)},
				}).Times(1)
				mockMutableState.EXPECT().GetCurrentBranchToken().Return(nil, nil).Times(1)
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{DomainID: "testDomain"}).Times(1)
				mockMutableState.EXPECT().GetNextEventID().Return(int64(10)).Times(1)
				mockShard.EXPECT().GetShardID().Return(1).Times(1)
				mockShard.EXPECT().GetDomainCache().Return(domainCache).Times(1)
				domainCache.EXPECT().GetDomainName("testDomain").Return("testDomain", nil).Times(1)

				eventTypeTimerFired := types.EventTypeTimerFired
				eventTypeTimerCanceled := types.EventTypeTimerCanceled
				events := []*types.HistoryEvent{
					{EventType: &eventTypeTimerFired, TimerFiredEventAttributes: &types.TimerFiredEventAttributes{TimerID: "timer1"}},
					{EventType: &eventTypeTimerCanceled, TimerCanceledEventAttributes: &types.TimerCanceledEventAttributes{TimerID: "timer2"}},
				}

				mockShard.EXPECT().GetHistoryManager().Return(manager).Times(1)
				manager.EXPECT().GetHistoryTree(gomock.Any(), gomock.Any()).Return(&persistence.GetHistoryTreeResponse{
					Branches: []*workflow.HistoryBranch{
						{TreeID: common.StringPtr("treeID"), BranchID: common.StringPtr("branchID")},
					}}, nil).AnyTimes()

				manager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Any()).Return(&persistence.ReadHistoryBranchResponse{
					HistoryEvents: events,
				}, nil).Times(1)
			},
			want: map[string]struct{}{"timer1": {}, "timer2": {}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			mockShard := shard.NewMockContext(mockCtrl)
			mockMutableState := NewMockMutableState(mockCtrl)
			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			mockHistoryManager := persistence.NewMockHistoryManager(mockCtrl)

			tc.setup(mockShard, mockMutableState, mockDomainCache, mockHistoryManager)
			ctx := context.Background()
			got, err := GetResurrectedTimers(ctx, mockShard, mockMutableState)

			if tc.wantErr {
				assert.Error(t, err, "GetResurrectedTimers() should have returned an error")
			} else {
				assert.NoError(t, err, "GetResurrectedTimers() should not have returned an error")
				assert.Equal(t, tc.want, got, "Mismatch in expected and actual resurrected timers")
			}
		})
	}
}

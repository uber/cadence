// The MIT License (MIT)
//
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

package fetcher

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
)

func TestGetUserTimers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRetryer := persistence.NewMockRetryer(ctrl)
	ctx := context.Background()
	pageSize := 10
	minTimestamp := time.Now().Add(-time.Hour)
	maxTimestamp := time.Now()

	testCases := []struct {
		name          string
		setupMock     func() []*persistence.TimerTaskInfo
		expectedPage  pagination.Page
		expectedError bool
	}{
		{
			name: "Success",
			setupMock: func() []*persistence.TimerTaskInfo {
				timerTasks := []*persistence.TimerTaskInfo{
					{
						DomainID:            "testDomainID",
						WorkflowID:          "testWorkflowID",
						RunID:               "testRunID",
						VisibilityTimestamp: time.Now(),
						TaskType:            persistence.TaskTypeUserTimer,
					},
				}

				mockRetryer.EXPECT().
					GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
						MinTimestamp:  minTimestamp,
						MaxTimestamp:  maxTimestamp,
						BatchSize:     pageSize,
						NextPageToken: nil,
					}).
					Return(&persistence.GetTimerIndexTasksResponse{
						Timers:        timerTasks,
						NextPageToken: nil,
					}, nil)

				mockRetryer.EXPECT().GetShardID().Return(123)

				return timerTasks
			},
			expectedPage: pagination.Page{
				Entities: []pagination.Entity{
					&entity.Timer{
						ShardID:             123,
						DomainID:            "testDomainID",
						WorkflowID:          "testWorkflowID",
						RunID:               "testRunID",
						TaskType:            persistence.TaskTypeUserTimer,
						VisibilityTimestamp: time.Now(), // This will be set to the current time in the test
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			timerTasks := tc.setupMock()
			fetchFn := getUserTimers(mockRetryer, minTimestamp, maxTimestamp, pageSize)
			page, err := fetchFn(ctx, nil)

			// Update the expected timestamp to match the mock data
			if len(timerTasks) > 0 {
				tc.expectedPage.Entities[0].(*entity.Timer).VisibilityTimestamp = timerTasks[0].VisibilityTimestamp
			}

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPage, page)
			}
		})
	}
}

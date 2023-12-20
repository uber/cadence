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
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
)

func TestGetUserTimers(t *testing.T) {
	fixedTimestamp, err := time.Parse(time.RFC3339, "2023-12-12T22:08:41Z")
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	ctrl := gomock.NewController(t)

	mockRetryer := persistence.NewMockRetryer(ctrl)
	ctx := context.Background()
	pageSize := 10
	minTimestamp := fixedTimestamp.Add(-time.Hour)
	maxTimestamp := fixedTimestamp
	nonNilToken := []byte("non-nil-token")
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
						VisibilityTimestamp: fixedTimestamp,
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
						VisibilityTimestamp: fixedTimestamp,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Non-nil Pagination Token Passed",
			setupMock: func() []*persistence.TimerTaskInfo {
				nonNilToken := []byte("non-nil-token")

				mockRetryer.EXPECT().
					GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
						MinTimestamp:  minTimestamp,
						MaxTimestamp:  maxTimestamp,
						BatchSize:     pageSize,
						NextPageToken: nonNilToken,
					}).
					Return(&persistence.GetTimerIndexTasksResponse{
						Timers:        nil, // Adjust based on expected behavior
						NextPageToken: nonNilToken,
					}, nil)

				return nil // No timers expected in response
			},
			expectedPage: pagination.Page{
				Entities:     nil,
				CurrentToken: nonNilToken,
				NextToken:    nonNilToken,
			},
			expectedError: false,
		},
		{
			name: "Invalid Timer Causes Error",
			setupMock: func() []*persistence.TimerTaskInfo {
				// Example setup of an invalid timer
				invalidTimer := &persistence.TimerTaskInfo{
					DomainID:            "",               // Invalid as it's empty
					WorkflowID:          "testWorkflowID", // Valid
					RunID:               "testRunID",      // Valid
					VisibilityTimestamp: time.Time{},      // Invalid as it's the zero value
					TaskType:            persistence.TaskTypeUserTimer,
				}

				mockRetryer.EXPECT().
					GetTimerIndexTasks(ctx, &persistence.GetTimerIndexTasksRequest{
						MinTimestamp:  minTimestamp,
						MaxTimestamp:  maxTimestamp,
						BatchSize:     pageSize,
						NextPageToken: nil,
					}).
					Return(&persistence.GetTimerIndexTasksResponse{
						Timers:        []*persistence.TimerTaskInfo{invalidTimer},
						NextPageToken: nil,
					}, nil)

				mockRetryer.EXPECT().GetShardID().Return(123)

				return []*persistence.TimerTaskInfo{invalidTimer}
			},
			expectedPage:  pagination.Page{}, // No page is expected due to error
			expectedError: true,              // An error is expected due to the invalid timer
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			var token pagination.PageToken
			if tc.name == "Non-nil Pagination Token Passed" {
				token = nonNilToken
			}
			fetchFn := getUserTimers(mockRetryer, minTimestamp, maxTimestamp, pageSize)
			page, err := fetchFn(ctx, token)

			require.Equal(t, tc.expectedError, err != nil)
			require.True(t, reflect.DeepEqual(tc.expectedPage.CurrentToken, page.CurrentToken), "CurrentToken should match")
			require.True(t, reflect.DeepEqual(tc.expectedPage.NextToken, page.NextToken), "NextToken should match")
			require.Equal(t, tc.expectedPage.Entities, page.Entities, "Entities should match")
		})
	}
}

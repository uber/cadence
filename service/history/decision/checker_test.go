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

package decision

import (
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/maps"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
)

const (
	testDomainID   = "test-domain-id"
	testDomainName = "test-domain"
	testWorkflowID = "test-workflow-id"
	testRunID      = "test-run-id"
)

func TestWorkflowSizeChecker_failWorkflowIfBlobSizeExceedsLimit(t *testing.T) {
	var (
		testDecisionTag = metrics.DecisionTypeTag(types.DecisionTypeCompleteWorkflowExecution.String())
		testEventID     = int64(1)
		testMessage     = "test"
	)

	for name, tc := range map[string]struct {
		blobSizeLimitWarn    int
		blobSizeLimitError   int
		blob                 []byte
		assertLogsAndMetrics func(*testing.T, *observer.ObservedLogs, tally.TestScope)
		expectFail           bool
	}{
		"no errors": {
			blobSizeLimitWarn:  10,
			blobSizeLimitError: 20,
			blob:               []byte("test"),
			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				assert.Empty(t, logs.All())
				// ensure metrics with the size is emitted.
				timerData := maps.Values(scope.Snapshot().Timers())
				assert.Len(t, timerData, 1)
				assert.Equal(t, "test.event_blob_size", timerData[0].Name())
			},
		},
		"warn": {
			blobSizeLimitWarn:  10,
			blobSizeLimitError: 20,
			blob:               []byte("should-warn"),
			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "Blob size exceeds limit.", logEntries[0].Message)
			},
		},
		"fail": {
			blobSizeLimitWarn:  5,
			blobSizeLimitError: 10,
			blob:               []byte("should-fail"),
			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "Blob size exceeds limit.", logEntries[0].Message)
			},
			expectFail: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mutableState := execution.NewMockMutableState(ctrl)
			logger, logs := testlogger.NewObserved(t)
			metricsScope := tally.NewTestScope("test", nil)
			checker := &workflowSizeChecker{
				blobSizeLimitWarn:  tc.blobSizeLimitWarn,
				blobSizeLimitError: tc.blobSizeLimitError,
				completedID:        testEventID,
				mutableState:       mutableState,
				logger:             logger,
				metricsScope:       metrics.NewClient(metricsScope, metrics.History).Scope(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.DomainTag(testDomainName)),
			}
			mutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
				DomainID:   testDomainID,
				WorkflowID: testWorkflowID,
				RunID:      testRunID,
			}).Times(1)
			if tc.expectFail {
				mutableState.EXPECT().AddFailWorkflowEvent(testEventID, &types.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr(common.FailureReasonDecisionBlobSizeExceedsLimit),
					Details: []byte(testMessage),
				}).Return(nil, nil).Times(1)
			}
			failed, err := checker.failWorkflowIfBlobSizeExceedsLimit(testDecisionTag, tc.blob, testMessage)
			require.NoError(t, err)
			if tc.assertLogsAndMetrics != nil {
				tc.assertLogsAndMetrics(t, logs, metricsScope)
			}
			assert.Equal(t, tc.expectFail, failed)
		})
	}

}

func TestWorkflowSizeChecker_failWorkflowSizeExceedsLimit(t *testing.T) {
	for name, tc := range map[string]struct {
		historyCount           int
		historyCountLimitWarn  int
		historyCountLimitError int

		historySize           int
		historySizeLimitWarn  int
		historySizeLimitError int

		noExecutionCall bool

		assertLogsAndMetrics func(*testing.T, *observer.ObservedLogs, tally.TestScope)
		expectFail           bool
	}{
		"no errors": {
			historyCount:           1,
			historyCountLimitWarn:  10,
			historyCountLimitError: 20,
			historySize:            1,
			historySizeLimitWarn:   10,
			historySizeLimitError:  20,
			noExecutionCall:        true,
			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				assert.Empty(t, logs.All())
				// ensure metrics with the size is emitted.
				timerData := maps.Values(scope.Snapshot().Timers())
				assert.Len(t, timerData, 4)
				timerNames := make([]string, 0, 4)
				for _, timer := range timerData {
					timerNames = append(timerNames, timer.Name())
				}
				sort.Strings(timerNames)

				// timers are duplicated for specific domain and domain: all
				assert.Equal(t, []string{"test.history_count", "test.history_count", "test.history_size", "test.history_size"}, timerNames)
			},
		},
		"count warn": {
			historyCount:           15,
			historyCountLimitWarn:  10,
			historyCountLimitError: 20,

			historySize:           1,
			historySizeLimitWarn:  10,
			historySizeLimitError: 20,

			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "history size exceeds warn limit.", logEntries[0].Message)
			},
		},
		"count error": {
			historyCount:           25,
			historyCountLimitWarn:  10,
			historyCountLimitError: 20,

			historySize:           1,
			historySizeLimitWarn:  10,
			historySizeLimitError: 20,

			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "history size exceeds error limit.", logEntries[0].Message)
			},
			expectFail: true,
		},
		"size warn": {
			historyCount:           1,
			historyCountLimitWarn:  10,
			historyCountLimitError: 20,

			historySize:           15,
			historySizeLimitWarn:  10,
			historySizeLimitError: 20,

			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "history size exceeds warn limit.", logEntries[0].Message)
			},
		},
		"size error": {
			historyCount:           1,
			historyCountLimitWarn:  10,
			historyCountLimitError: 20,

			historySize:           25,
			historySizeLimitWarn:  10,
			historySizeLimitError: 20,

			assertLogsAndMetrics: func(t *testing.T, logs *observer.ObservedLogs, scope tally.TestScope) {
				logEntries := logs.All()
				require.Len(t, logEntries, 1)
				assert.Equal(t, "history size exceeds error limit.", logEntries[0].Message)
			},
			expectFail: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mutableState := execution.NewMockMutableState(ctrl)
			logger, logs := testlogger.NewObserved(t)
			metricsScope := tally.NewTestScope("test", nil)

			mutableState.EXPECT().GetNextEventID().Return(int64(tc.historyCount + 1)).Times(1)
			if !tc.noExecutionCall {
				mutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
					DomainID:   testDomainID,
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				}).Times(1)
			}

			checker := &workflowSizeChecker{
				historyCountLimitWarn:  tc.historyCountLimitWarn,
				historyCountLimitError: tc.historyCountLimitError,
				historySizeLimitWarn:   tc.historySizeLimitWarn,
				historySizeLimitError:  tc.historySizeLimitError,
				mutableState:           mutableState,
				executionStats: &persistence.ExecutionStats{
					HistorySize: int64(tc.historySize),
				},
				logger:       logger,
				metricsScope: metrics.NewClient(metricsScope, metrics.History).Scope(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.DomainTag(testDomainName)),
			}
			failed, err := checker.failWorkflowSizeExceedsLimit()
			require.NoError(t, err)
			if tc.assertLogsAndMetrics != nil {
				tc.assertLogsAndMetrics(t, logs, metricsScope)
			}
			assert.Equal(t, tc.expectFail, failed)
		})
	}
}

// Copyright (c) 2017 Uber Technologies, Inc.
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

// Code generated by thriftrw v1.19.1. DO NOT EDIT.
// @generated

package sqlblobs

import (
	shared "github.com/uber/cadence/.gen/go/shared"
	thriftreflect "go.uber.org/thriftrw/thriftreflect"
)

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "sqlblobs",
	Package:  "github.com/uber/cadence/.gen/go/sqlblobs",
	FilePath: "sqlblobs.thrift",
	SHA1:     "4e175efebe843be1ae2668ae45d28873a3c77492",
	Includes: []*thriftreflect.ThriftModule{
		shared.ThriftModule,
	},
	Raw: rawIDL,
}

const rawIDL = "// Copyright (c) 2017 Uber Technologies, Inc.\n//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\nnamespace java com.uber.cadence.sqlblobs\n\ninclude \"shared.thrift\"\n\nstruct ShardInfo {\n  10: optional i32 stolenSinceRenew\n  12: optional i64 (js.type = \"Long\") updatedAtNanos\n  14: optional i64 (js.type = \"Long\") replicationAckLevel\n  16: optional i64 (js.type = \"Long\") transferAckLevel\n  18: optional i64 (js.type = \"Long\") timerAckLevelNanos\n  24: optional i64 (js.type = \"Long\") domainNotificationVersion\n  34: optional map<string, i64> clusterTransferAckLevel\n  36: optional map<string, i64> clusterTimerAckLevel\n  38: optional string owner\n}\n\nstruct DomainInfo {\n  10: optional string name\n  12: optional string description\n  14: optional string owner\n  16: optional i32 status\n  18: optional i16 retentionDays\n  20: optional bool emitMetric\n  22: optional string archivalBucket\n  24: optional i16 archivalStatus\n  26: optional i64 (js.type = \"Long\") configVersion\n  28: optional i64 (js.type = \"Long\") notificationVersion\n  30: optional i64 (js.type = \"Long\") failoverNotificationVersion\n  32: optional i64 (js.type = \"Long\") failoverVersion\n  34: optional string activeClusterName\n  36: optional list<string> clusters\n  38: optional map<string, string> data\n  39: optional binary badBinaries\n  40: optional string badBinariesEncoding\n}\n\nstruct HistoryTreeInfo {\n  10: optional i64 (js.type = \"Long\") createdTimeNanos // For fork operation to prevent race condition of leaking event data when forking branches fail. Also can be used for clean up leaked data\n  12: optional list<shared.HistoryBranchRange> ancestors\n  14: optional string info // For lookup back to workflow during debugging, also background cleanup when fork operation cannot finish self cleanup due to crash.\n}\n\nstruct ReplicationInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional i64 (js.type = \"Long\") lastEventID\n}\n\nstruct WorkflowExecutionInfo {\n  10: optional binary parentDomainID\n  12: optional string parentWorkflowID\n  14: optional binary parentRunID\n  16: optional i64 (js.type = \"Long\") initiatedID\n  18: optional i64 (js.type = \"Long\") completionEventBatchID\n  20: optional binary completionEvent\n  22: optional string completionEventEncoding\n  24: optional string taskList\n  26: optional string workflowTypeName\n  28: optional i32 workflowTimeoutSeconds\n  30: optional i32 decisionTaskTimeoutSeconds\n  32: optional binary executionContext\n  34: optional i32 state\n  36: optional i32 closeStatus\n  38: optional i64 (js.type = \"Long\") startVersion\n  40: optional i64 (js.type = \"Long\") currentVersion\n  44: optional i64 (js.type = \"Long\") lastWriteEventID\n  46: optional map<string, ReplicationInfo> lastReplicationInfo\n  48: optional i64 (js.type = \"Long\") lastEventTaskID\n  50: optional i64 (js.type = \"Long\") lastFirstEventID\n  52: optional i64 (js.type = \"Long\") lastProcessedEvent\n  54: optional i64 (js.type = \"Long\") startTimeNanos\n  56: optional i64 (js.type = \"Long\") lastUpdatedTimeNanos\n  58: optional i64 (js.type = \"Long\") decisionVersion\n  60: optional i64 (js.type = \"Long\") decisionScheduleID\n  62: optional i64 (js.type = \"Long\") decisionStartedID\n  64: optional i32 decisionTimeout\n  66: optional i64 (js.type = \"Long\") decisionAttempt\n  68: optional i64 (js.type = \"Long\") decisionStartedTimestampNanos\n  69: optional i64 (js.type = \"Long\") decisionScheduledTimestampNanos\n  70: optional bool cancelRequested\n  72: optional string createRequestID\n  74: optional string decisionRequestID\n  76: optional string cancelRequestID\n  78: optional string stickyTaskList\n  80: optional i64 (js.type = \"Long\") stickyScheduleToStartTimeout\n  82: optional i64 (js.type = \"Long\") retryAttempt\n  84: optional i32 retryInitialIntervalSeconds\n  86: optional i32 retryMaximumIntervalSeconds\n  88: optional i32 retryMaximumAttempts\n  90: optional i32 retryExpirationSeconds\n  92: optional double retryBackoffCoefficient\n  94: optional i64 (js.type = \"Long\") retryExpirationTimeNanos\n  96: optional list<string> retryNonRetryableErrors\n  98: optional bool hasRetryPolicy\n  100: optional string cronSchedule\n  102: optional i32 eventStoreVersion\n  104: optional binary eventBranchToken\n  106: optional i64 (js.type = \"Long\") signalCount\n  108: optional i64 (js.type = \"Long\") historySize\n  110: optional string clientLibraryVersion\n  112: optional string clientFeatureVersion\n  114: optional string clientImpl\n  115: optional binary autoResetPoints\n  116: optional string autoResetPointsEncoding\n  118: optional map<string, binary> searchAttributes\n}\n\nstruct ActivityInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional i64 (js.type = \"Long\") scheduledEventBatchID\n  14: optional binary scheduledEvent\n  16: optional string scheduledEventEncoding\n  18: optional i64 (js.type = \"Long\") scheduledTimeNanos\n  20: optional i64 (js.type = \"Long\") startedID\n  22: optional binary startedEvent\n  24: optional string startedEventEncoding\n  26: optional i64 (js.type = \"Long\") startedTimeNanos\n  28: optional string activityID\n  30: optional string requestID\n  32: optional i32 scheduleToStartTimeoutSeconds\n  34: optional i32 scheduleToCloseTimeoutSeconds\n  36: optional i32 startToCloseTimeoutSeconds\n  38: optional i32 heartbeatTimeoutSeconds\n  40: optional bool cancelRequested\n  42: optional i64 (js.type = \"Long\") cancelRequestID\n  44: optional i32 timerTaskStatus\n  46: optional i32 attempt\n  48: optional string taskList\n  50: optional string startedIdentity\n  52: optional bool hasRetryPolicy\n  54: optional i32 retryInitialIntervalSeconds\n  56: optional i32 retryMaximumIntervalSeconds\n  58: optional i32 retryMaximumAttempts\n  60: optional i64 (js.type = \"Long\") retryExpirationTimeNanos\n  62: optional double retryBackoffCoefficient\n  64: optional list<string> retryNonRetryableErrors\n}\n\nstruct ChildExecutionInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional i64 (js.type = \"Long\") initiatedEventBatchID\n  14: optional i64 (js.type = \"Long\") startedID\n  16: optional binary initiatedEvent\n  18: optional string initiatedEventEncoding\n  20: optional string startedWorkflowID\n  22: optional binary startedRunID\n  24: optional binary startedEvent\n  26: optional string startedEventEncoding\n  28: optional string createRequestID\n  30: optional string domainName\n  32: optional string workflowTypeName\n}\n\nstruct SignalInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional string requestID\n  14: optional string name\n  16: optional binary input\n  18: optional binary control\n}\n\nstruct RequestCancelInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional string cancelRequestID\n}\n\nstruct TimerInfo {\n  10: optional i64 (js.type = \"Long\") version\n  12: optional i64 (js.type = \"Long\") startedID\n  14: optional i64 (js.type = \"Long\") expiryTimeNanos\n  16: optional i64 (js.type = \"Long\") taskID\n}\n\nstruct TaskInfo {\n  10: optional string workflowID\n  12: optional binary runID\n  13: optional i64 (js.type = \"Long\") scheduleID\n  14: optional i64 (js.type = \"Long\") expiryTimeNanos\n  15: optional i64 (js.type = \"Long\") createdTimeNanos\n}\n\nstruct TaskListInfo {\n  10: optional i16 kind // {Normal, Sticky}\n  12: optional i64 (js.type = \"Long\") ackLevel\n  14: optional i64 (js.type = \"Long\") expiryTimeNanos\n  16: optional i64 (js.type = \"Long\") lastUpdatedNanos\n}\n\nstruct TransferTaskInfo {\n  10: optional binary domainID\n  12: optional string workflowID\n  14: optional binary runID\n  16: optional i16 taskType\n  18: optional binary targetDomainID\n  20: optional string targetWorkflowID\n  22: optional binary targetRunID\n  24: optional string taskList\n  26: optional bool targetChildWorkflowOnly\n  28: optional i64 (js.type = \"Long\") scheduleID\n  30: optional i64 (js.type = \"Long\") version\n  32: optional i64 (js.type = \"Long\") visibilityTimestampNanos\n}\n\nstruct TimerTaskInfo {\n  10: optional binary domainID\n  12: optional string workflowID\n  14: optional binary runID\n  16: optional i16 taskType\n  18: optional i16 timeoutType\n  20: optional i64 (js.type = \"Long\") version\n  22: optional i64 (js.type = \"Long\") scheduleAttempt\n  24: optional i64 (js.type = \"Long\") eventID\n}\n\nstruct ReplicationTaskInfo {\n  10: optional binary domainID\n  12: optional string workflowID\n  14: optional binary runID\n  16: optional i16 taskType\n  18: optional i64 (js.type = \"Long\") version\n  20: optional i64 (js.type = \"Long\") firstEventID\n  22: optional i64 (js.type = \"Long\") nextEventID\n  24: optional i64 (js.type = \"Long\") scheduledID\n  26: optional i32 eventStoreVersion\n  28: optional i32 newRunEventStoreVersion\n  30: optional binary branch_token\n  32: optional map<string, ReplicationInfo> lastReplicationInfo\n  34: optional binary newRunBranchToken\n  36: optional bool resetWorkflow\n}"

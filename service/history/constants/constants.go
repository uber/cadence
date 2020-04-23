// Copyright (c) 2020 Uber Technologies, Inc.
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

package constants

const (
	// WorkflowTerminationReason is the reason for terminating workflow due to version conflit
	WorkflowTerminationReason = "Terminate Workflow Due To Version Conflict."
	// WorkflowTerminationIdentity is the component which decides to terminate the workflow
	WorkflowTerminationIdentity = "worker-service"
	// ErrRetrySyncActivityMsg is returned when sync activity replication tasks are arriving out of order, should retry
	ErrRetrySyncActivityMsg = "retry on applying sync activity"
	// EventsReapplicationResetWorkflowReason is the reason for reset workflow during reapplication
	EventsReapplicationResetWorkflowReason = "events-reapplication"
	// IdentityHistoryService is the service role identity
	IdentityHistoryService = "history-service"
	// NDCDefaultPageSize is the default pagination size for ndc
	NDCDefaultPageSize = 100
)

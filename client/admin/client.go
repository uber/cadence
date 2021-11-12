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

package admin

import "time"

var (
	// DefaultTimeout is the default timeout used to make calls
	DefaultTimeout = 10 * time.Second
	// DefaultLargeTimeout is the default timeout used to make calls
	DefaultLargeTimeout = time.Minute

	// MaxTimeouts specify a max allowed duration for each method on the client.
	// It it used to override context deadline, shortening it when it is too far in the future.
	// If the value is nil, additional max timeout is not enforced and current context deadline is untouched.
	// We use it as a safeguard to prevent service exhaustion, when upstream timeout is too large.
	MaxTimeouts = map[string]*time.Duration{
		"AddSearchAttribute":                &DefaultTimeout,
		"DescribeShardDistribution":         &DefaultTimeout,
		"DescribeHistoryHost":               &DefaultTimeout,
		"RemoveTask":                        &DefaultTimeout,
		"CloseShard":                        &DefaultTimeout,
		"ResetQueue":                        &DefaultTimeout,
		"DescribeQueue":                     &DefaultTimeout,
		"DescribeWorkflowExecution":         &DefaultTimeout,
		"GetWorkflowExecutionRawHistoryV2":  &DefaultTimeout,
		"DescribeCluster":                   &DefaultTimeout,
		"GetReplicationMessages":            &DefaultLargeTimeout,
		"GetDomainReplicationMessages":      &DefaultTimeout,
		"GetDLQReplicationMessages":         &DefaultTimeout,
		"ReapplyEvents":                     &DefaultTimeout,
		"ReadDLQMessages":                   &DefaultTimeout,
		"PurgeDLQMessages":                  &DefaultTimeout,
		"MergeDLQMessages":                  &DefaultTimeout,
		"RefreshWorkflowTasks":              &DefaultTimeout,
		"ResendReplicationTasks":            &DefaultTimeout,
		"GetCrossClusterTasks":              &DefaultLargeTimeout,
		"RespondCrossClusterTasksCompleted": &DefaultTimeout,
		"GetDynamicConfig":                  &DefaultTimeout,
		"UpdateDynamicConfig":               &DefaultTimeout,
		"RestoreDynamicConfig":              &DefaultTimeout,
		"ListDynamicConfig":                 &DefaultTimeout,
	}
)

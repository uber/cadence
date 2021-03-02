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

package shadower

import (
	"errors"
	"time"

	"go.uber.org/cadence/workflow"
)

// The following constants should be identical to those defined in the client repo
// TODO: figure out a way to keep those constants in sync

const (
	scanWorkflowActivityName            = "scanWorkflowActivity"
	replayWorkflowExecutionActivityName = "replayWorkflowExecutionActivity"
)

const (
	errMsgDomainNotExists           = "domain not exists"
	errMsgInvalidQuery              = "invalid visibility query"
	errMsgWorkflowTypeNotRegistered = "workflow type not registered"
)

type (
	shadowWorkflowParams struct {
		Version int

		Domain   string
		TaskList string

		WorkflowQuery string
		NextPageToken []byte
		SamplingRate  float64

		ShadowMode    shadowMode
		ExitCondition exitCondition

		Concurrency int

		// ActivityConfigs activityConfigs
	}

	// activityConfigs struct {
	// 	ScanWorkflowActivityName   string
	// 	ReplayWorkflowActivityName string
	// 	NonRetryableErrors         []string
	// }

	shadowWorkflowResult struct {
		Succeed        int
		Skipped        int
		Failed         int
		FailureDetails []string // keep last x failure information
	}

	shadowMode int

	exitCondition struct {
		ExpirationTime time.Duration
		ShadowingCount int
	}
)

const (
	shadowModeNormal shadowMode = iota + 1
	shadowModeContinuous
)

func shadowWorkflow(
	ctx workflow.Context,
	params shadowWorkflowParams,
) (shadowWorkflowResult, error) {
	return shadowWorkflowResult{}, errors.New("TODO: implement shadow workflow")
}

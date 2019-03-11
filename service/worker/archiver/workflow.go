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

package archiver

import (
	"time"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/logging"
	"go.uber.org/cadence/workflow"
)

func archivalWorkflow(ctx workflow.Context, carryover []ArchiveRequest) error {
	workflowInfo := workflow.GetInfo(ctx)
	globalLogger.WithFields(bark.Fields{
		logging.TagDomain:              workflowInfo.Domain,
		logging.TagWorkflowExecutionID: workflowInfo.WorkflowExecution.ID,
		logging.TagWorkflowRunID:       workflowInfo.WorkflowExecution.RunID,
	}).Info("archival system workflow started")
	config, err := readConfig(ctx)
	if err != nil {
		// log and emit metric
		return err
	}
	requestCh := workflow.NewBufferedChannel(ctx, config.ArchivalsPerIteration)
	archiver := NewProcessor(ctx, globalLogger, globalMetricsClient, config.ProcessorConcurrency, requestCh)
	archiver.Start()
	signalCh := workflow.GetSignalChannel(ctx, signalName)
	pump := NewPump(ctx, globalLogger, globalMetricsClient, carryover, workflowStartToCloseTimeout/2, config.ArchivalsPerIteration, requestCh, signalCh)
	pumpResult := pump.Run()
	handledHashes := archiver.Finished()
	if pumpResult.TimeoutWithoutSignals {
		return nil
	}
	if !equal(pumpResult.PumpedHashes, handledHashes) {
		// log and emit metric
	}
	for {
		var request ArchiveRequest
		if ok := signalCh.ReceiveAsync(&request); !ok {
			break
		}
		pumpResult.UnhandledCarryover = append(pumpResult.UnhandledCarryover, request)
	}
	ctx = workflow.WithExecutionStartToCloseTimeout(ctx, workflowStartToCloseTimeout)
	ctx = workflow.WithWorkflowTaskStartToCloseTimeout(ctx, workflowTaskStartToCloseTimeout)
	return workflow.NewContinueAsNewError(ctx, archivalWorkflowFnName, pumpResult.UnhandledCarryover)
}

func readConfig(ctx workflow.Context) (readConfigActivityResult, error) {
	opts := workflow.ActivityOptions{
		ScheduleToStartTimeout: 30 * time.Second,
		StartToCloseTimeout:    30 * time.Second,
	}
	actCtx := workflow.WithActivityOptions(ctx, opts)
	var result readConfigActivityResult
	if err := workflow.ExecuteActivity(actCtx, readConfigActivityFnName).Get(actCtx, &result); err != nil {
		return readConfigActivityResult{}, err
	}
	return result, nil
}

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
	"go.uber.org/cadence/workflow"
)

func archivalWorkflow(ctx workflow.Context, carryover []ArchiveRequest) error {
	archivalsPerIteration := globalConfig.ArchivalsPerIteration() // TODO: consider if reading dynamic config needs to be in activity
	requestCh := workflow.NewBufferedChannel(ctx, archivalsPerIteration)
	archiver := NewArchiver(ctx, globalLogger, globalMetricsClient, globalConfig.ArchiverConcurrency(), requestCh)
	archiver.Start()
	signalCh := workflow.GetSignalChannel(ctx, signalName)
	pump := NewPump(ctx, globalLogger, globalMetricsClient, carryover, workflowStartToCloseTimeout / 2, archivalsPerIteration, requestCh, signalCh)
	pumpResult := pump.Run()
	handledHashes := archiver.Finished()
	if pumpResult.TimeoutWithoutSignals {
		// TODO: emit metric in this case
		return nil
	}
	if len(pumpResult.UnhandledCarryover) > 0 {
		// TODO: should emit metric here and log indicating that backlog is building
	}
	if !Equal(pumpResult.PumpedHashes, handledHashes) {
		// TODO: should emit metric here and log indicating that something very bad happened
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
	return workflow.NewContinueAsNewError(ctx, archiveSystemWorkflowFnName, pumpResult.UnhandledCarryover)
}

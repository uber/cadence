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
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
)

// TODO: rename sysworkflow to archival, this package is specific to triggering and doing archival



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











//// ArchiveSystemWorkflow is the system workflow which archives and deletes history
//func ArchiveSystemWorkflow2(ctx workflow.Context, carryoverRequests []ArchiveRequest) error {
//	sysWorkflowInfo := workflow.GetInfo(ctx)
//	logger := NewReplayBarkLogger(globalLogger.WithFields(bark.Fields{
//		logging.TagWorkflowExecutionID: sysWorkflowInfo.WorkflowExecution.ID,
//		logging.TagWorkflowRunID:       sysWorkflowInfo.WorkflowExecution.RunID,
//	}), ctx, false)
//	logger.Info("started system workflow")
//
//	metricsClient := NewReplayMetricsClient(globalMetricsClient, ctx)
//	metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerWorkflowStarted)
//	sw := metricsClient.StartTimer(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerContinueAsNewLatency)
//	requestsHandled := 0
//
//	// step 1: start workers to process archival requests in parallel
//	workQueue := workflow.NewChannel(ctx)
//	finishedWorkQueue := workflow.NewBufferedChannel(ctx, signalsUntilContinueAsNew*10) // make large enough that never blocks on send
//	for i := 0; i < numWorkers; i++ {
//		workflow.Go(ctx, func(ctx workflow.Context) {
//			for {
//				var request ArchiveRequest
//				workQueue.Receive(ctx, &request)
//				// handleRequest(request, ctx, logger, metricsClient)
//				finishedWorkQueue.Send(ctx, nil)
//			}
//		})
//	}
//
//	// step 2: pump carryover requests into worker queue
//	for _, request := range carryoverRequests {
//		requestsHandled++
//		workQueue.Send(ctx, request)
//	}
//
//	// step 3: pump current iterations workload into worker queue
//	ch := workflow.GetSignalChannel(ctx, signalName)
//	for requestsHandled < signalsUntilContinueAsNew {
//		var request ArchiveRequest
//		if more := ch.Receive(ctx, &request); !more {
//			break
//		}
//		metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerReceivedSignal)
//		requestsHandled++
//		workQueue.Send(ctx, request)
//	}
//
//	// step 4: wait for all in progress work to finish
//	for i := 0; i < requestsHandled; i++ {
//		finishedWorkQueue.Receive(ctx, nil)
//	}
//
//	// step 5: drain signal channel to get next run's carryover
//	var co []ArchiveRequest
//	for {
//		var request ArchiveRequest
//		if ok := ch.ReceiveAsync(&request); !ok {
//			break
//		}
//		metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerReceivedSignal)
//		co = append(co, request)
//	}
//
//	// step 6: schedule new run
//	ctx = workflow.WithExecutionStartToCloseTimeout(ctx, workflowStartToCloseTimeout)
//	ctx = workflow.WithWorkflowTaskStartToCloseTimeout(ctx, decisionTaskStartToCloseTimeout)
//	metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerContinueAsNew)
//	sw.Stop()
//	logger.WithFields(bark.Fields{
//		logging.TagNumberOfSignalsUntilContinueAsNew: requestsHandled,
//	}).Info("system workflow is continuing as new")
//	return workflow.NewContinueAsNewError(ctx, archiveSystemWorkflowFnName, co)
//}


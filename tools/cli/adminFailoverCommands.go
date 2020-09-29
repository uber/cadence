// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package cli

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	cclient "go.uber.org/cadence/client"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/service/worker/failoverManager"
)

const (
	defaultAbortReason                    = "Failover aborted through admin CLI"
	defaultBatchFailoverSize              = 10
	defaultBatchFailoverWaitTimeInSeconds = 10
	defaultFailoverTimeoutInSeconds       = 600
)

type startParams struct {
	targetCluster                  string
	sourceCluster                  string
	batchFailoverSize              int
	batchFailoverWaitTimeInSeconds int
	failoverTimeout                int
	domains                        []string
}

func failoverStart(c *cli.Context, params *startParams) {
	validateStartParams(params)

	targetCluster := params.targetCluster
	sourceCluster := params.sourceCluster
	batchFailoverSize := params.batchFailoverSize
	batchFailoverWaitTimeInSeconds := params.batchFailoverWaitTimeInSeconds
	workflowTimeout := time.Duration(params.failoverTimeout) * time.Second
	domains := params.domains

	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	options := cclient.StartWorkflowOptions{
		ID:                           failoverManager.WorkflowID,
		WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
		TaskList:                     failoverManager.TaskListName,
		ExecutionStartToCloseTimeout: workflowTimeout,
	}
	foParams := failoverManager.FailoverParams{
		TargetCluster:                  targetCluster,
		SourceCluster:                  sourceCluster,
		BatchFailoverSize:              batchFailoverSize,
		BatchFailoverWaitTimeInSeconds: batchFailoverWaitTimeInSeconds,
		Domains:                        domains,
	}
	wf, err := client.StartWorkflow(tcCtx, options, failoverManager.WorkflowTypeName, foParams)
	if err != nil {
		ErrorAndExit("Failed to start failover workflow", err)
	}
	fmt.Println("Failover workflow started")
	fmt.Println("wid: " + wf.ID)
	fmt.Println("rid: " + wf.RunID)
}

func validateStartParams(params *startParams) {
	if len(params.targetCluster) == 0 {
		ErrorAndExit("targetCluster is not provided", nil)
	}
	if len(params.sourceCluster) == 0 {
		ErrorAndExit("sourceCluster is not provided", nil)
	}
	if params.targetCluster == params.sourceCluster {
		ErrorAndExit("targetCluster is same as sourceCluster", nil)
	}
	if params.batchFailoverSize <= 0 {
		params.batchFailoverSize = defaultBatchFailoverSize
	}
	if params.batchFailoverWaitTimeInSeconds <= 0 {
		params.batchFailoverWaitTimeInSeconds = defaultBatchFailoverWaitTimeInSeconds
	}
	if params.failoverTimeout <= 0 {
		params.failoverTimeout = defaultFailoverTimeoutInSeconds
	}
}

// AdminFailoverStart start failover workflow
func AdminFailoverStart(c *cli.Context) {
	params := &startParams{
		targetCluster:                  getRequiredOption(c, FlagTargetCluster),
		sourceCluster:                  getRequiredOption(c, FlagSourceCluster),
		batchFailoverSize:              c.Int(FlagFailoverBatchSize),
		batchFailoverWaitTimeInSeconds: c.Int(FlagFailoverWaitTime),
		failoverTimeout:                c.Int(FlagFailoverTimeout),
		domains:                        c.StringSlice(FlagFailoverDomains),
	}
	failoverStart(c, params)
}

// AdminFailoverPause pause failover workflow
func AdminFailoverPause(c *cli.Context) {
	err := executePauseOrResume(c, true)
	if err != nil {
		ErrorAndExit("Failed to pause failover workflow", err)
	}
	fmt.Println("Failover paused")
}

// AdminFailoverResume resume a paused failover workflow
func AdminFailoverResume(c *cli.Context) {
	err := executePauseOrResume(c, false)
	if err != nil {
		ErrorAndExit("Failed to resume failover workflow", err)
	}
	fmt.Println("Failover resumed")
}

func executePauseOrResume(c *cli.Context, isPause bool) error {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	runID := getRunID(c, client)
	var signalName string
	if isPause {
		signalName = failoverManager.PauseSignal
	} else {
		signalName = failoverManager.ResumeSignal
	}

	return client.SignalWorkflow(tcCtx, failoverManager.WorkflowID, runID, signalName, nil)
}

// AdminFailoverQuery query a failover workflow
func AdminFailoverQuery(c *cli.Context) {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	runID := getRunID(c, client)
	queryResult, err := client.QueryWorkflow(tcCtx, failoverManager.WorkflowID, runID, failoverManager.QueryType)
	if err != nil {
		ErrorAndExit("Failed to query failover workflow", err)
	}
	if !queryResult.HasValue() {
		ErrorAndExit("QueryResult has no value", nil)
	}
	var result failoverManager.QueryResult
	queryResult.Get(&result)
	prettyPrintJSONObject(result)
}

// AdminFailoverAbort abort a failover workflow
func AdminFailoverAbort(c *cli.Context) {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	reason := c.String(FlagReason)
	if len(reason) == 0 {
		reason = defaultAbortReason
	}
	runID := getRunID(c, client)

	// todo: update query state to reflect abort status
	err := client.TerminateWorkflow(tcCtx, failoverManager.WorkflowID, runID, reason, nil)
	if err != nil {
		ErrorAndExit("Failed to abort failover workflow", err)
	}

	fmt.Println("Failover aborted")
}

// AdminFailoverRollback rollback a failover run
func AdminFailoverRollback(c *cli.Context) {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	runID := getRunID(c, client)

	queryResp, err := client.QueryWorkflow(tcCtx, failoverManager.WorkflowID, runID, failoverManager.QueryType)
	if err != nil {
		ErrorAndExit("Failed to query failover workflow", err)
	}
	if !queryResp.HasValue() {
		ErrorAndExit("QueryResult has no value", nil)
	}
	var queryResult failoverManager.QueryResult
	queryResp.Get(&queryResult)

	params := &startParams{
		targetCluster:                  queryResult.SourceCluster,
		sourceCluster:                  queryResult.TargetCluster,
		domains:                        queryResult.SuccessDomains,
		batchFailoverSize:              c.Int(FlagFailoverBatchSize),
		batchFailoverWaitTimeInSeconds: c.Int(FlagFailoverWaitTime),
		failoverTimeout:                c.Int(FlagFailoverTimeout),
	}
	failoverStart(c, params)
}

func getCadenceClient(c *cli.Context) cclient.Client {
	svcClient := cFactory.ClientFrontendClient(c)
	return cclient.NewClient(svcClient, common.SystemLocalDomainName, &cclient.Options{})
}

func getRunID(c *cli.Context, client cclient.Client) string {
	var runID string

	if c.IsSet(FlagRunID) {
		runID = c.String(FlagRunID)
	} else {
		tcCtx, cancel := newContext(c)
		defer cancel()

		descResp, err := client.DescribeWorkflowExecution(tcCtx, failoverManager.WorkflowID, "")
		if err != nil {
			ErrorAndExit("Failed to get failover workflow runID", err)
		}

		runID = descResp.WorkflowExecutionInfo.Execution.GetRunId()
	}
	return runID
}

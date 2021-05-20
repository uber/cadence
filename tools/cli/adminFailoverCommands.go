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
	"context"
	"fmt"
	"os/user"
	"time"

	"go.uber.org/cadence/.gen/go/shared"

	"github.com/urfave/cli"
	cclient "go.uber.org/cadence/client"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/service/worker/failovermanager"
)

const (
	defaultAbortReason                    = "Failover aborted through admin CLI"
	defaultBatchFailoverSize              = 20
	defaultBatchFailoverWaitTimeInSeconds = 30
	defaultFailoverTimeoutInSeconds       = 1200
)

type startParams struct {
	targetCluster                  string
	sourceCluster                  string
	batchFailoverSize              int
	batchFailoverWaitTimeInSeconds int
	failoverTimeout                int
	domains                        []string
	drillWaitTime                  int
	cron                           string
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
		drillWaitTime:                  c.Int(FlagFailoverDrillWaitTime),
		cron:                           c.String(FlagCronSchedule),
	}
	failoverStart(c, params)
}

// AdminFailoverPause pause failover workflow
func AdminFailoverPause(c *cli.Context) {
	err := executePauseOrResume(c, getFailoverWorkflowID(c), true)
	if err != nil {
		ErrorAndExit("Failed to pause failover workflow", err)
	}
	fmt.Println("Failover paused on " + getFailoverWorkflowID(c))
}

// AdminFailoverResume resume a paused failover workflow
func AdminFailoverResume(c *cli.Context) {
	err := executePauseOrResume(c, getFailoverWorkflowID(c), false)
	if err != nil {
		ErrorAndExit("Failed to resume failover workflow", err)
	}
	fmt.Println("Failover resumed on " + getFailoverWorkflowID(c))
}

// AdminFailoverQuery query a failover workflow
func AdminFailoverQuery(c *cli.Context) {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()
	workflowID := getFailoverWorkflowID(c)
	runID := getRunID(c)
	result := query(tcCtx, client, workflowID, runID)

	descResp, err := client.DescribeWorkflowExecution(tcCtx, workflowID, runID)
	if err != nil {
		ErrorAndExit("Failed to describe workflow", err)
	}
	if isWorkflowTerminated(descResp) {
		result.State = failovermanager.WorkflowAborted
	}
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
	workflowID := getFailoverWorkflowID(c)
	runID := getRunID(c)

	err := client.TerminateWorkflow(tcCtx, workflowID, runID, reason, nil)
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

	runID := getRunID(c)

	queryResult := query(tcCtx, client, failovermanager.FailoverWorkflowID, runID)
	if isWorkflowRunning(queryResult) {
		err := client.TerminateWorkflow(tcCtx, failovermanager.FailoverWorkflowID, runID, "Rollback", nil)
		if err != nil {
			ErrorAndExit("Failed to terminate failover workflow", err)
		}
	}
	// query again
	queryResult = query(tcCtx, client, failovermanager.FailoverWorkflowID, runID)
	var rollbackDomains []string
	// rollback includes both success and failed domains to make sure no leftover domains
	rollbackDomains = append(rollbackDomains, queryResult.SuccessDomains...)
	rollbackDomains = append(rollbackDomains, queryResult.FailedDomains...)

	params := &startParams{
		targetCluster:                  queryResult.SourceCluster,
		sourceCluster:                  queryResult.TargetCluster,
		domains:                        rollbackDomains,
		batchFailoverSize:              c.Int(FlagFailoverBatchSize),
		batchFailoverWaitTimeInSeconds: c.Int(FlagFailoverWaitTime),
		failoverTimeout:                c.Int(FlagFailoverTimeout),
	}
	failoverStart(c, params)
}

// AdminFailoverList list failover runs
func AdminFailoverList(c *cli.Context) {
	c.Set(FlagWorkflowID, getFailoverWorkflowID(c))
	c.GlobalSet(FlagDomain, common.SystemLocalDomainName)
	ListWorkflow(c)
}

func query(
	tcCtx context.Context,
	client cclient.Client,
	workflowID string,
	runID string) *failovermanager.QueryResult {
	queryResp, err := client.QueryWorkflow(tcCtx, workflowID, runID, failovermanager.QueryType)
	if err != nil {
		ErrorAndExit("Failed to query failover workflow", err)
	}
	if !queryResp.HasValue() {
		ErrorAndExit("QueryResult has no value", nil)
	}
	var queryResult failovermanager.QueryResult
	queryResp.Get(&queryResult)
	return &queryResult
}

func isWorkflowRunning(queryResult *failovermanager.QueryResult) bool {
	return queryResult.State == failovermanager.WorkflowRunning ||
		queryResult.State == failovermanager.WorkflowPaused
}

func getCadenceClient(c *cli.Context) cclient.Client {
	svcClient := cFactory.ClientFrontendClient(c)
	return cclient.NewClient(svcClient, common.SystemLocalDomainName, &cclient.Options{})
}

func getRunID(c *cli.Context) string {
	if c.IsSet(FlagRunID) {
		return c.String(FlagRunID)
	}
	return ""
}

func failoverStart(c *cli.Context, params *startParams) {
	validateStartParams(params)

	targetCluster := params.targetCluster
	sourceCluster := params.sourceCluster
	batchFailoverSize := params.batchFailoverSize
	batchFailoverWaitTimeInSeconds := params.batchFailoverWaitTimeInSeconds
	workflowTimeout := time.Duration(params.failoverTimeout) * time.Second
	domains := params.domains
	drillWaitTime := time.Duration(params.drillWaitTime) * time.Second

	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	options := cclient.StartWorkflowOptions{
		ID:                           failovermanager.FailoverWorkflowID,
		WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
		TaskList:                     failovermanager.TaskListName,
		ExecutionStartToCloseTimeout: workflowTimeout,
		Memo: map[string]interface{}{
			common.MemoKeyForOperator: getOperator(),
		},
	}
	if params.drillWaitTime > 0 {
		options.ID = failovermanager.DrillWorkflowID
		options.CronSchedule = params.cron
	} else {

		if len(params.cron) > 0 {
			ErrorAndExit("The drill wait time is required when cron is specified.", nil)
		}

		// block if there is an on-going failover drill
		if err := executePauseOrResume(c, failovermanager.DrillWorkflowID, true); err != nil {
			switch err.(type) {
			case *shared.EntityNotExistsError:
				break
			case *shared.WorkflowExecutionAlreadyCompletedError:
				break
			default:
				ErrorAndExit("Failed to send pase signal to drill workflow", err)
			}
		}
		fmt.Println("The failover drill workflow is paused. Please run 'cadence admin cluster failover resume --fd'" +
			" to resume the drill workflow.")
	}

	foParams := failovermanager.FailoverParams{
		TargetCluster:                  targetCluster,
		SourceCluster:                  sourceCluster,
		BatchFailoverSize:              batchFailoverSize,
		BatchFailoverWaitTimeInSeconds: batchFailoverWaitTimeInSeconds,
		Domains:                        domains,
		DrillWaitTime:                  drillWaitTime,
	}
	wf, err := client.StartWorkflow(tcCtx, options, failovermanager.FailoverWorkflowTypeName, foParams)
	if err != nil {
		ErrorAndExit("Failed to start failover workflow", err)
	}
	fmt.Println("Failover workflow started")
	fmt.Println("wid: " + wf.ID)
	fmt.Println("rid: " + wf.RunID)
}

func getFailoverWorkflowID(c *cli.Context) string {
	if c.Bool(FlagFailoverDrill) {
		return failovermanager.DrillWorkflowID
	}
	return failovermanager.FailoverWorkflowID
}

func getOperator() string {
	user, err := user.Current()
	if err != nil {
		ErrorAndExit("Unable to get operator info", err)
	}

	return fmt.Sprintf("%s (username: %s)", user.Name, user.Username)
}

func isWorkflowTerminated(descResp *shared.DescribeWorkflowExecutionResponse) bool {
	return descResp.WorkflowExecutionInfo.CloseStatus != nil &&
		descResp.WorkflowExecutionInfo.CloseStatus.Equals(shared.WorkflowExecutionCloseStatusTerminated)
}

func executePauseOrResume(c *cli.Context, workflowID string, isPause bool) error {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	runID := getRunID(c)
	var signalName string
	if isPause {
		signalName = failovermanager.PauseSignal
	} else {
		signalName = failovermanager.ResumeSignal
	}

	return client.SignalWorkflow(tcCtx, workflowID, runID, signalName, nil)
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

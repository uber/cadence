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

// AdminFailoverStart start failover workflow
func AdminFailoverStart(c *cli.Context) {
	targetCluster := getRequiredOption(c, FlagTargetCluster)
	batchFailoverSize := c.Int(FlagFailoverBatchSize)
	batchFailoverWaitTimeInSeconds := c.Int(FlagFailoverWaitTime)
	workflowTimeout := time.Duration(c.Int(FlagFailoverTimeout)) * time.Second

	svcClient := cFactory.ClientFrontendClient(c)
	client := cclient.NewClient(svcClient, common.SystemLocalDomainName, &cclient.Options{})
	tcCtx, cancel := newContext(c)
	defer cancel()

	options := cclient.StartWorkflowOptions{
		ID:                           failoverManager.WorkflowID,
		WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
		TaskList:                     failoverManager.TaskListName,
		ExecutionStartToCloseTimeout: workflowTimeout,
	}
	params := failoverManager.FailoverParams{
		TargetCluster:                  targetCluster,
		BatchFailoverSize:              batchFailoverSize,
		BatchFailoverWaitTimeInSeconds: batchFailoverWaitTimeInSeconds,
	}
	wf, err := client.StartWorkflow(tcCtx, options, failoverManager.WorkflowTypeName, params)
	if err != nil {
		ErrorAndExit("Failed to start failover workflow", err)
	}
	fmt.Println("Failover workflow started")
	fmt.Println("wid: " + wf.ID)
	fmt.Println("rid: " + wf.RunID)
}

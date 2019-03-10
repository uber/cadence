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
	"context"
	"fmt"
	"github.com/uber/cadence/common"
	"math/rand"

	"github.com/uber/cadence/client/public"
	"github.com/uber/cadence/common/service/dynamicconfig"
	cclient "go.uber.org/cadence/client"
)

// TODO: regenerate client_mock once code builds again...

type (
	// ArchiveRequest is request to Archive
	ArchiveRequest struct {
		DomainID             string
		WorkflowID           string
		RunID                string
		EventStoreVersion    int32
		BranchToken          []byte
		NextEventID          int64
		CloseFailoverVersion int64
	}

	// Client is used to archive workflow histories
	Client interface {
		Archive(*ArchiveRequest) error
	}

	client struct {
		cadenceClient cclient.Client
		numWorkflows  dynamicconfig.IntPropertyFn
	}
)

// NewClient creates a new Client
func NewClient(publicClient public.Client, numWorkflows dynamicconfig.IntPropertyFn) Client {
	return &client{
		cadenceClient: cclient.NewClient(publicClient, common.SystemDomainName, &cclient.Options{}),
		numWorkflows:  numWorkflows,
	}
}

// Archive starts an archival task
func (c *client) Archive(request *ArchiveRequest) error {
	workflowID := fmt.Sprintf("%v-%v", workflowIDPrefix, rand.Intn(c.numWorkflows()))
	workflowOptions := cclient.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        decisionTaskList,
		ExecutionStartToCloseTimeout:    workflowStartToCloseTimeout,
		DecisionTaskStartToCloseTimeout: workflowTaskStartToCloseTimeout,
		WorkflowIDReusePolicy:           cclient.WorkflowIDReusePolicyAllowDuplicate,
	}
	_, err := c.cadenceClient.SignalWithStartWorkflow(context.Background(), workflowID, signalName, *request, workflowOptions, archivalWorkflowFnName, nil)
	return err
}

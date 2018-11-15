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

package sysworkflow

import (
	"context"
	"fmt"
	"github.com/uber/cadence/client/frontend"
	"go.uber.org/cadence/client"
	"time"
)

type (
	Initiator interface {
		Archive(request *ArchiveRequest) error
		// Add any other system tasks you want here...
	}
	initiator struct {
		cadenceClient  client.Client
		sysWorkflowIDs []string
	}
	Signal struct {
		RequestType    RequestType
		ArchiveRequest *ArchiveRequest
	}
	ArchiveRequest struct {
		UserWorkflowID string
		UserRunID      string
	}
)

func NewInitiator(frontendClient frontend.Client, numSysWorkflows int) Initiator {
	return &initiator{
		cadenceClient:  client.NewClient(frontendClient, Domain, &client.Options{}),
		sysWorkflowIDs: RandomSlice(numSysWorkflows),
	}
}
func (i *initiator) Archive(request *ArchiveRequest) error {
	workflowID := PickRandom(i.sysWorkflowIDs)
	workflowOptions := client.StartWorkflowOptions{
		ID: workflowID,
		// TODO: once we have higher load, this should select one random of X task lists to do load balancing
		TaskList:                        DecisionTaskList,
		ExecutionStartToCloseTimeout:    time.Hour * 48,
		DecisionTaskStartToCloseTimeout: time.Minute * 10,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
	signal := Signal{
		RequestType:    ArchivalRequest,
		ArchiveRequest: request,
	}
	fmt.Println("sending signal")
	_, err := i.cadenceClient.SignalWithStartWorkflow(
		context.Background(),
		workflowID,
		SignalName,
		signal,
		workflowOptions,
		SystemWorkflow,
	)
	return err
}

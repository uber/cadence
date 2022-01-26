// Copyright (c) 2022 Uber Technologies, Inc.
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

package watchdog

import (
	"context"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	cclient "go.uber.org/cadence/client"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (

	// Client is used to send request to processor workflow
	Client interface {
		ReportCorruptWorkflow(domainName string, workflowID string, runID string) error
	}

	clientImpl struct {
		metricsClient metrics.Client
		logger        log.Logger
		cadenceClient cclient.Client
	}
)

var _ Client = (*clientImpl)(nil)

const (
	signalTimeout = 400 * time.Millisecond
)

// NewClient creates a new Client
func NewClient(
	metricsClient metrics.Client,
	logger log.Logger,
	publicClient workflowserviceclient.Interface,
) Client {
	return &clientImpl{
		metricsClient: metricsClient,
		logger:        logger,
		cadenceClient: cclient.NewClient(publicClient, common.SystemLocalDomainName, &cclient.Options{}),
	}
}

func (c *clientImpl) ReportCorruptWorkflow(
	domainName string,
	workflowID string,
	runID string,
) error {
	signalCtx, cancel := context.WithTimeout(context.Background(), signalTimeout)
	defer cancel()
	request := CorruptWFRequest{
		DomainName: domainName,
		workflow: types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	}
	err := c.cadenceClient.SignalWorkflow(signalCtx, watchdogWFID, "", corruptWorkflowWatchdogChannelName, request)
	return err
}

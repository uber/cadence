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

package esanalyzer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/workercommon"
)

type (
	// Analyzer is the background sub-system to query ElasticSearch and execute mitigations
	Analyzer struct {
		svcClient           workflowserviceclient.Interface
		frontendClient      frontend.Client
		esClient            es.GenericClient
		logger              log.Logger
		metricsClient       metrics.Client
		tallyScope          tally.Scope
		visibilityIndexName string
		resource            resource.Resource
	}
)

// New returns a new instance as daemon
func New(
	svcClient workflowserviceclient.Interface,
	frontendClient frontend.Client,
	esClient es.GenericClient,
	esConfig *config.ElasticSearchConfig,
	logger log.Logger,
	metricsClient metrics.Client,
	tallyScope tally.Scope,
	resource resource.Resource,
) *Analyzer {
	return &Analyzer{
		svcClient:           svcClient,
		frontendClient:      frontendClient,
		esClient:            esClient,
		logger:              logger,
		metricsClient:       metricsClient,
		tallyScope:          tallyScope,
		visibilityIndexName: esConfig.Indices[common.VisibilityAppName],
		resource:            resource,
	}
}

// Start starts the scanner
func (a *Analyzer) Start() error {
	ctx := context.WithValue(context.Background(), analyzerContextKey, a)

	a.StartWorkflow(ctx)

	workerOpts := worker.Options{
		MetricsScope:              a.tallyScope,
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	esWorker := worker.New(a.svcClient, common.SystemLocalDomainName, taskListName, workerOpts)
	err := esWorker.Start()
	return err
}

func (a *Analyzer) StartWorkflow(ctx context.Context) {
	go workercommon.StartWorkflowWithRetry(wfTypeName, startUpDelay, a.resource, func(client client.Client) error {
		_, err := client.StartWorkflow(ctx, wfOptions, wfTypeName)
		switch err.(type) {
		case *shared.WorkflowExecutionAlreadyStartedError:
			return nil
		default:
			return err
		}
	})
}

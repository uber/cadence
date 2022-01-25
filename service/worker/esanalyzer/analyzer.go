// Copyright (c) 2021 Uber Technologies, Inc.
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
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/workercommon"
)

type (
	// Analyzer is the background sub-system to query ElasticSearch and execute mitigations
	Analyzer struct {
		svcClient           workflowserviceclient.Interface
		frontendClient      frontend.Client
		clientBean          client.Bean
		esClient            es.GenericClient
		logger              log.Logger
		scopedMetricClient  metrics.Scope
		tallyScope          tally.Scope
		visibilityIndexName string
		resource            resource.Resource
		domainCache         cache.DomainCache
		config              *Config
	}

	// Config contains all configs for ElasticSearch Analyzer
	Config struct {
		ESAnalyzerPause                          dynamicconfig.BoolPropertyFn
		ESAnalyzerTimeWindow                     dynamicconfig.DurationPropertyFn
		ESAnalyzerMaxNumDomains                  dynamicconfig.IntPropertyFn
		ESAnalyzerMaxNumWorkflowTypes            dynamicconfig.IntPropertyFn
		ESAnalyzerLimitToTypes                   dynamicconfig.StringPropertyFn
		ESAnalyzerLimitToDomains                 dynamicconfig.StringPropertyFn
		ESAnalyzerNumWorkflowsToRefresh          dynamicconfig.IntPropertyFnWithWorkflowTypeFilter
		ESAnalyzerBufferWaitTime                 dynamicconfig.DurationPropertyFnWithWorkflowTypeFilter
		ESAnalyzerMinNumWorkflowsForAvg          dynamicconfig.IntPropertyFnWithWorkflowTypeFilter
		ESAnalyzerWorkflowDurationWarnThresholds dynamicconfig.StringPropertyFn
	}
)

const startUpDelay = time.Second * 10

// New returns a new instance as daemon
func New(
	svcClient workflowserviceclient.Interface,
	frontendClient frontend.Client,
	clientBean client.Bean,
	esClient es.GenericClient,
	esConfig *config.ElasticSearchConfig,
	logger log.Logger,
	metricsClient metrics.Client,
	tallyScope tally.Scope,
	resource resource.Resource,
	domainCache cache.DomainCache,
	config *Config,
) *Analyzer {
	return &Analyzer{
		svcClient:           svcClient,
		frontendClient:      frontendClient,
		clientBean:          clientBean,
		esClient:            esClient,
		logger:              logger,
		scopedMetricClient:  getScopedMetricsClient(metricsClient),
		tallyScope:          tallyScope,
		visibilityIndexName: esConfig.Indices[common.VisibilityAppName],
		resource:            resource,
		domainCache:         domainCache,
		config:              config,
	}
}

func getScopedMetricsClient(metricsClient metrics.Client) metrics.Scope {
	return metricsClient.Scope(metrics.ESAnalyzerScope)
}

// Start starts the scanner
func (a *Analyzer) Start() error {
	ctx := context.Background()
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
	initWorkflow(a)
	go workercommon.StartWorkflowWithRetry(esanalyzerWFTypeName, startUpDelay, a.resource, func(client cclient.Client) error {
		_, err := client.StartWorkflow(ctx, wfOptions, esanalyzerWFTypeName)
		switch err.(type) {
		case *shared.WorkflowExecutionAlreadyStartedError:
			return nil
		default:
			a.logger.Error("Failed to start ElasticSearch Analyzer", tag.Error(err))
			return err
		}
	})
}

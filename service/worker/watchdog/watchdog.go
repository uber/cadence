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
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/workercommon"
)

type (
	// WatchDog is the background sub-system to maintain the health of a Cadence cluster
	WatchDog struct {
		svcClient          workflowserviceclient.Interface
		frontendClient     frontend.Client
		clientBean         client.Bean
		logger             log.Logger
		scopedMetricClient metrics.Scope
		tallyScope         tally.Scope
		resource           resource.Resource
		domainCache        cache.DomainCache
		config             *Config
	}

	// Config contains all configs for ElasticSearch WatchDog
	Config struct {
		CorruptWorkflowPause dynamicconfig.BoolPropertyFn
	}
)

const startUpDelay = time.Second * 10

// New returns a new instance as daemon
func New(
	svcClient workflowserviceclient.Interface,
	frontendClient frontend.Client,
	clientBean client.Bean,
	logger log.Logger,
	metricsClient metrics.Client,
	tallyScope tally.Scope,
	resource resource.Resource,
	domainCache cache.DomainCache,
	config *Config,
) *WatchDog {
	return &WatchDog{
		svcClient:          svcClient,
		frontendClient:     frontendClient,
		clientBean:         clientBean,
		logger:             logger,
		scopedMetricClient: getScopedMetricsClient(metricsClient),
		tallyScope:         tallyScope,
		resource:           resource,
		domainCache:        domainCache,
		config:             config,
	}
}

func getScopedMetricsClient(metricsClient metrics.Client) metrics.Scope {
	return metricsClient.Scope(metrics.WatchDogScope)
}

// Start starts the scanner
func (wd *WatchDog) Start() error {
	ctx := context.Background()
	wd.StartWorkflow(ctx)

	workerOpts := worker.Options{
		MetricsScope:              wd.tallyScope,
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	esWorker := worker.New(wd.svcClient, common.SystemLocalDomainName, taskListName, workerOpts)
	err := esWorker.Start()
	return err
}

func (wd *WatchDog) StartWorkflow(ctx context.Context) {
	initWorkflow(wd)
	go workercommon.StartWorkflowWithRetry(watchdogWFTypeName, startUpDelay, wd.resource, func(client cclient.Client) error {
		_, err := client.StartWorkflow(ctx, wfOptions, watchdogWFTypeName)
		switch err.(type) {
		case *shared.WorkflowExecutionAlreadyStartedError:
			return nil
		default:
			wd.logger.Error("Failed to start WatchDog", tag.Error(err))
			return err
		}
	})
}

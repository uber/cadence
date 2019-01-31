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
	"github.com/uber-common/bark"
	"github.com/uber/cadence/client/public"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

type (
	// Config for SysWorker
	Config struct {
		EnableArchivalCompression dynamicconfig.BoolPropertyFn
	}
	// SysWorker is the cadence client worker responsible for running system workflows
	SysWorker struct {
		worker worker.Worker
	}
)

func init() {
	workflow.RegisterWithOptions(SystemWorkflow, workflow.RegisterOptions{Name: systemWorkflowFnName})
	activity.RegisterWithOptions(ArchivalActivity, activity.RegisterOptions{Name: archivalActivityFnName})
	activity.RegisterWithOptions(BackfillActivity, activity.RegisterOptions{Name: backfillActivityFnName})
}

// NewSysWorker returns a new SysWorker
func NewSysWorker(
	publicClient public.Client,
	metrics metrics.Client,
	logger bark.Logger,
	clusterMetadata cluster.Metadata,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	blobstore blobstore.Client,
	domainCache cache.DomainCache,
	config *Config) *SysWorker {

	actCtx := context.Background()
	actCtx = context.WithValue(actCtx, publicClientKey, publicClient)
	actCtx = context.WithValue(actCtx, metricsKey, metrics)
	actCtx = context.WithValue(actCtx, loggerKey, logger)
	actCtx = context.WithValue(actCtx, clusterMetadataKey, clusterMetadata)
	actCtx = context.WithValue(actCtx, historyManagerKey, historyManager)
	actCtx = context.WithValue(actCtx, historyV2ManagerKey, historyV2Manager)
	actCtx = context.WithValue(actCtx, blobstoreKey, blobstore)
	actCtx = context.WithValue(actCtx, domainCacheKey, domainCache)
	actCtx = context.WithValue(actCtx, configKey, config)
	wo := worker.Options{
		BackgroundActivityContext: actCtx,
	}
	return &SysWorker{
		worker: worker.New(publicClient, SystemDomainName, decisionTaskList, wo),
	}
}

// Start the SysWorker
func (w *SysWorker) Start() error {
	if err := w.worker.Start(); err != nil {
		w.worker.Stop()
		return err
	}
	return nil
}

// Stop the SysWorker
func (w *SysWorker) Stop() {
	w.worker.Stop()
}

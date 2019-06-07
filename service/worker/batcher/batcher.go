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

package batcher

import (
	"context"
	"time"

	"github.com/uber/cadence/common/service/dynamicconfig"

	"github.com/uber-go/tally"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/worker"
)

const (
	// maximum time waiting for this batcher to start before giving up
	maxTimeForStartup = time.Minute * 3
)

type (
	// Config defines the configuration for batcher
	Config struct {
		AdminOperationToken dynamicconfig.StringPropertyFn
		// ClusterMetadata contains the metadata for this cluster
		ClusterMetadata cluster.Metadata
	}

	// BootstrapParams contains the set of params needed to bootstrap
	// the batcher sub-system
	BootstrapParams struct {
		// Config contains the configuration for scanner
		Config Config
		// ServiceClient is an instance of cadence service client
		ServiceClient workflowserviceclient.Interface
		// MetricsClient is an instance of metrics object for emitting stats
		MetricsClient metrics.Client
		Logger        log.Logger
		// TallyScope is an instance of tally metrics scope
		TallyScope tally.Scope
	}

	// batcherContext is the context object that get's
	// passed around within the scanner workflows / activities
	batcherContext struct {
		cfg           Config
		svcClient     workflowserviceclient.Interface
		metricsClient metrics.Client
		tallyScope    tally.Scope
		logger        log.Logger
	}

	// Batcher is the background sub-system that execute workflow for batch operations
	Batcher struct {
		context batcherContext
	}
)

// New returns a new instance of batcher daemon Batcher
func New(params *BootstrapParams) *Batcher {
	cfg := params.Config
	return &Batcher{
		context: batcherContext{
			cfg:           cfg,
			svcClient:     params.ServiceClient,
			metricsClient: params.MetricsClient,
			tallyScope:    params.TallyScope,
			logger:        params.Logger.WithTags(tag.ComponentBatcher),
		},
	}
}

// Start starts the scanner
func (s *Batcher) Start() error {
	workerOpts := worker.Options{
		MetricsScope:              s.context.tallyScope,
		BackgroundActivityContext: context.WithValue(context.Background(), batcherContextKey, s.context),
	}
	worker := worker.New(s.context.svcClient, common.SystemDomainName, tlBatcherTaskListName, workerOpts)
	err := worker.Start()
	if err != nil {
		return err
	}

	//retry until making sure global system domain is there
	return s.createGlobalSystemDomainIfNotExistsWithRetry()
}

func (s *Batcher) createGlobalSystemDomainIfNotExistsWithRetry() error {
	policy := backoff.NewExponentialRetryPolicy(time.Second)
	policy.SetMaximumInterval(time.Minute)
	policy.SetExpirationInterval(maxTimeForStartup)
	return backoff.Retry(func() error {
		return s.createGlobalSystemDomainIfNotExists()
	}, policy, func(err error) bool {
		return true
	})
}

func (s *Batcher) createGlobalSystemDomainIfNotExists() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := s.context.svcClient.DescribeDomain(ctx, &shared.DescribeDomainRequest{
		Name: common.StringPtr(common.SystemGlobalDomainName),
	})
	cancel()

	if err == nil {
		s.context.logger.Info("global system domain already exists", tag.WorkflowDomainName(common.SystemGlobalDomainName))
		return nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	err = s.context.svcClient.RegisterDomain(ctx, s.getDomainCreationRequest())
	cancel()
	if err != nil {
		s.context.logger.Error("error creating global system domain", tag.Error(err))
		return err
	}
	s.context.logger.Info("Domain created successfully")
	return nil
}

func (s *Batcher) getDomainCreationRequest() *shared.RegisterDomainRequest {
	req := &shared.RegisterDomainRequest{
		Name:                                   common.StringPtr(common.SystemGlobalDomainName),
		WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(common.SystemDomainRetentionDays),
		EmitMetric:                             common.BoolPtr(true),
		SecurityToken:                          common.StringPtr(s.context.cfg.AdminOperationToken()),
		IsGlobalDomain:                         common.BoolPtr(false),
	}

	if s.context.cfg.ClusterMetadata.IsGlobalDomainEnabled() {
		req.IsGlobalDomain = common.BoolPtr(true)
		req.ActiveClusterName = common.StringPtr(s.context.cfg.ClusterMetadata.GetMasterClusterName())
		var clusters []*shared.ClusterReplicationConfiguration
		for name := range s.context.cfg.ClusterMetadata.GetAllClusterInfo() {
			clusters = append(clusters, &shared.ClusterReplicationConfiguration{
				ClusterName: common.StringPtr(name),
			})
		}
		req.Clusters = clusters
	}

	return req
}

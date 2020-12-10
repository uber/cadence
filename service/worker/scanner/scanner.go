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

package scanner

import (
	"context"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/shared"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

const (
	// scannerStartUpDelay is to let services warm up
	scannerStartUpDelay = time.Second * 4
)

type (
	// Config defines the configuration for scanner
	Config struct {
		// ScannerPersistenceMaxQPS the max rate of calls to persistence
		// Right now is being used by historyScanner to determine the rate of persistence API calls
		ScannerPersistenceMaxQPS dynamicconfig.IntPropertyFn
		// Persistence contains the persistence configuration
		Persistence *config.Persistence
		// ClusterMetadata contains the metadata for this cluster
		ClusterMetadata cluster.Metadata
		// TaskListScannerEnabled indicates if taskList scanner should be started as part of scanner
		TaskListScannerEnabled dynamicconfig.BoolPropertyFn
		// HistoryScannerEnabled indicates if history scanner should be started as part of scanner
		HistoryScannerEnabled dynamicconfig.BoolPropertyFn
		// ShardScanners is a list of shard scanner configs
		ShardScanners []*shardscanner.ScannerConfig
	}

	// BootstrapParams contains the set of params needed to bootstrap
	// the scanner sub-system
	BootstrapParams struct {
		// Config contains the configuration for scanner
		Config Config
		// TallyScope is an instance of tally metrics scope
		TallyScope tally.Scope
	}

	// scannerContext is the context object that get's
	// passed around within the scanner workflows / activities
	scannerContext struct {
		resource.Resource
		cfg        Config
		tallyScope tally.Scope
		zapLogger  *zap.Logger
	}

	// Scanner is the background sub-system that does full scans
	// of database tables to cleanup resources, monitor anomalies
	// and emit stats for analytics
	Scanner struct {
		context scannerContext
	}
)

// New returns a new instance of scanner daemon
// Scanner is the background sub-system that does full
// scans of database tables in an attempt to cleanup
// resources, monitor system anamolies and emit stats
// for analysis and alerting
func New(
	resource resource.Resource,
	params *BootstrapParams,
) *Scanner {

	cfg := params.Config
	zapLogger, err := zap.NewProduction()
	if err != nil {
		resource.GetLogger().Fatal("failed to initialize zap logger", tag.Error(err))
	}
	return &Scanner{
		context: scannerContext{
			Resource:   resource,
			cfg:        cfg,
			tallyScope: params.TallyScope,
			zapLogger:  zapLogger.Named("scanner"),
		},
	}
}

// Start starts the scanner
func (s *Scanner) Start() error {
	backgroundActivityContext := context.Background()
	var workerTaskListNames []string
	var wtl []string

	for _, sc := range s.context.cfg.ShardScanners {
		s.context.GetLogger().Info("starting shard scanner", tag.WorkflowType(sc.ScannerWFTypeName))
		backgroundActivityContext, wtl = s.startShardScanner(backgroundActivityContext, sc)
		workerTaskListNames = append(workerTaskListNames, wtl...)
	}

	switch s.context.cfg.Persistence.DefaultStoreType() {
	case config.StoreTypeSQL:
		if s.context.cfg.TaskListScannerEnabled() {
			go s.startWorkflowWithRetry(tlScannerWFStartOptions, tlScannerWFTypeName, nil)
			workerTaskListNames = append(workerTaskListNames, tlScannerTaskListName)
		}
	case config.StoreTypeCassandra:
		if s.context.cfg.HistoryScannerEnabled() {
			go s.startWorkflowWithRetry(historyScannerWFStartOptions, historyScannerWFTypeName, nil)
			workerTaskListNames = append(workerTaskListNames, historyScannerTaskListName)
		}

	}

	workerOpts := worker.Options{
		Logger:                                 s.context.zapLogger,
		MetricsScope:                           s.context.tallyScope,
		MaxConcurrentActivityExecutionSize:     maxConcurrentActivityExecutionSize,
		MaxConcurrentDecisionTaskExecutionSize: maxConcurrentDecisionTaskExecutionSize,
		BackgroundActivityContext:              backgroundActivityContext,
	}

	for _, tl := range workerTaskListNames {
		if err := worker.New(s.context.GetSDKClient(), common.SystemLocalDomainName, tl, workerOpts).Start(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scanner) startShardScanner(
	ctx context.Context,
	config *shardscanner.ScannerConfig,
) (context.Context, []string) {
	scannerContextKey := shardscanner.ScannerContextKey(config.ScannerWFTypeName)
	fixerContextKey := shardscanner.ScannerContextKey(config.FixerWFTypeName)
	workerTaskListNames := []string{}
	backgroundActivityContext := context.WithValue(ctx, scannerContextKey, s.context)

	if config.DynamicParams.ScannerEnabled() {
		backgroundActivityContext = context.WithValue(
			backgroundActivityContext,
			scannerContextKey,
			getShardScannerContext(s.context, config),
		)
	}

	if config.DynamicParams.FixerEnabled() {
		backgroundActivityContext = context.WithValue(
			backgroundActivityContext,
			fixerContextKey,
			getShardFixerContext(s.context, config),
		)

		workerTaskListNames = append(workerTaskListNames, config.FixerTLName)
	}

	if config.DynamicParams.ScannerEnabled() {
		workerTaskListNames = append(workerTaskListNames, config.StartWorkflowOptions.TaskList)

		go s.startWorkflowWithRetry(config.StartWorkflowOptions, config.ScannerWFTypeName, shardscanner.ScannerWorkflowParams{
			Shards: shardscanner.Shards{
				Range: &shardscanner.ShardRange{
					Min: 0,
					Max: s.context.cfg.Persistence.NumHistoryShards,
				},
			},
		})
	}

	return backgroundActivityContext, workerTaskListNames
}

func (s *Scanner) startWorkflowWithRetry(
	options cclient.StartWorkflowOptions,
	workflowType string,
	workflowArg interface{},
) {

	// let history / matching service warm up
	time.Sleep(scannerStartUpDelay)

	sdkClient := cclient.NewClient(s.context.GetSDKClient(), common.SystemLocalDomainName, &cclient.Options{})
	policy := backoff.NewExponentialRetryPolicy(time.Second)
	policy.SetMaximumInterval(time.Minute)
	policy.SetExpirationInterval(backoff.NoInterval)
	err := backoff.Retry(func() error {
		return s.startWorkflow(sdkClient, options, workflowType, workflowArg)
	}, policy, func(err error) bool {
		return true
	})
	if err != nil {
		s.context.GetLogger().Fatal("unable to start scanner", tag.WorkflowType(workflowType), tag.Error(err))
	}
}

func (s *Scanner) startWorkflow(
	client cclient.Client,
	options cclient.StartWorkflowOptions,
	workflowType string,
	workflowArg interface{},
) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	var err error
	if workflowArg != nil {
		_, err = client.StartWorkflow(ctx, options, workflowType, workflowArg)
	} else {
		_, err = client.StartWorkflow(ctx, options, workflowType)
	}

	cancel()
	if err != nil {
		if _, ok := err.(*shared.WorkflowExecutionAlreadyStartedError); ok {
			return nil
		}
		s.context.GetLogger().Error("error starting workflow", tag.Error(err), tag.WorkflowType(workflowType))
		return err
	}
	s.context.GetLogger().Info("workflow successfully started", tag.WorkflowType(workflowType))
	return nil
}

func getShardScannerContext(ctx scannerContext, config *shardscanner.ScannerConfig) shardscanner.Context {
	return shardscanner.Context{
		Resource:   ctx.Resource,
		Scope:      ctx.Resource.GetMetricsClient().Scope(metrics.ExecutionsScannerScope),
		Config:     config,
		ContextKey: shardscanner.ScannerContextKey(config.ScannerWFTypeName),
		Hooks:      config.ScannerHooks(),
	}
}

func getShardFixerContext(ctx scannerContext, config *shardscanner.ScannerConfig) shardscanner.FixerContext {
	return shardscanner.FixerContext{
		Resource:   ctx.Resource,
		Scope:      ctx.Resource.GetMetricsClient().Scope(metrics.ExecutionsFixerScope),
		ContextKey: shardscanner.ScannerContextKey(config.FixerWFTypeName),
		Hooks:      config.FixerHooks(),
		Config:     config,
	}
}

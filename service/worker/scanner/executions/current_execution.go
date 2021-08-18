// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package executions

import (
	"context"
	"strconv"
	"time"

	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

const (
	// currentExecutionsScannerWFID is the current execution scanner workflow ID
	currentExecutionsScannerWFID = "cadence-sys-current-executions-scanner"
	// CurrentExecutionsScannerWFTypeName is the current execution scanner workflow type
	CurrentExecutionsScannerWFTypeName = "cadence-sys-current-executions-scanner-workflow"
	// CurrentExecutionsScannerTaskListName is the current execution scanner workflow tasklist
	CurrentExecutionsScannerTaskListName = "cadence-sys-current-executions-scanner-tasklist-0"

	// CurrentExecutionsFixerWFTypeName is the current execution fixer workflow ID
	CurrentExecutionsFixerWFTypeName = "cadence-sys-current-executions-fixer-workflow"
	currentExecutionsFixerWFID       = "cadence-sys-current-executions-fixer"
	// CurrentExecutionsFixerTaskListName is the current execution fixer workflow tasklist
	CurrentExecutionsFixerTaskListName = "cadence-sys-current-executions-fixer-tasklist-0"
)

// CurrentScannerWorkflow is the workflow that scans over all current executions
func CurrentScannerWorkflow(
	ctx workflow.Context,
	params shardscanner.ScannerWorkflowParams,
) error {
	wf, err := shardscanner.NewScannerWorkflow(ctx, CurrentExecutionsScannerWFTypeName, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

// CurrentExecutionsHooks provides hooks for current executions scanner.
func CurrentExecutionsHooks() *shardscanner.ScannerHooks {
	wf, err := shardscanner.NewScannerHooks(CurrentExecutionManager, CurrentExecutionIterator)
	if err != nil {
		return nil
	}
	wf.SetConfig(CurrentExecutionConfig)

	return wf
}

// CurrentExecutionManager is the current execution scanner manager
func CurrentExecutionManager(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
) invariant.Manager {
	var ivs []invariant.Invariant
	collections := ParseCollections(params.ScannerConfig)
	for _, fn := range CurrentExecutionType.ToInvariants(collections) {
		ivs = append(ivs, fn(pr))
	}
	return invariant.NewInvariantManager(ivs)
}

// CurrentFixerWorkflow starts current executions fixer.
func CurrentFixerWorkflow(
	ctx workflow.Context,
	params shardscanner.FixerWorkflowParams,
) error {
	wf, err := shardscanner.NewFixerWorkflow(ctx, CurrentExecutionsFixerWFTypeName, params)
	if err != nil {
		return err
	}
	return wf.Start(ctx)
}

// CurrentExecutionConfig resolves dynamic config for current executions scanner.
func CurrentExecutionConfig(ctx shardscanner.Context) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}

	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerInvariantCollectionHistory, true)() {
		res[invariant.CollectionHistory.String()] = strconv.FormatBool(true)
	}
	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerInvariantCollectionMutableState, true)() {
		res[invariant.CollectionMutableState.String()] = strconv.FormatBool(true)
	}

	return res
}

// CurrentExecutionFixerHooks provides hooks for current executions fixer.
func CurrentExecutionFixerHooks() *shardscanner.FixerHooks {
	h, err := shardscanner.NewFixerHooks(FixerManager, CurrentExecutionFixerIterator)
	if err != nil {
		return nil
	}
	return h
}

// CurrentExecutionScannerConfig configures current execution scanner
func CurrentExecutionScannerConfig(dc *dynamicconfig.Collection) *shardscanner.ScannerConfig {
	return &shardscanner.ScannerConfig{
		ScannerWFTypeName: CurrentExecutionsScannerWFTypeName,
		FixerWFTypeName:   CurrentExecutionsFixerWFTypeName,
		DynamicCollection: dc,
		DynamicParams: shardscanner.DynamicParams{
			ScannerEnabled:          dc.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerEnabled, false),
			FixerEnabled:            dc.GetBoolProperty(dynamicconfig.CurrentExecutionFixerEnabled, false),
			Concurrency:             dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerConcurrency, 25),
			PageSize:                dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerPersistencePageSize, 1000),
			BlobstoreFlushThreshold: dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerBlobstoreFlushThreshold, 100),
			ActivityBatchSize:       dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerActivityBatchSize, 25),
			AllowDomain:             dc.GetBoolPropertyFilteredByDomain(dynamicconfig.CurrentExecutionFixerDomainAllow, false),
		},
		ScannerHooks: CurrentExecutionsHooks,
		FixerHooks:   CurrentExecutionFixerHooks,
		StartWorkflowOptions: cclient.StartWorkflowOptions{
			ID:                           currentExecutionsScannerWFID,
			TaskList:                     CurrentExecutionsScannerTaskListName,
			ExecutionStartToCloseTimeout: 20 * 365 * 24 * time.Hour,
			WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
			CronSchedule:                 "* * * * *",
		},
		StartFixerOptions: cclient.StartWorkflowOptions{
			ID:                           currentExecutionsFixerWFID,
			TaskList:                     CurrentExecutionsFixerTaskListName,
			ExecutionStartToCloseTimeout: 20 * 365 * 24 * time.Hour,
			WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
			CronSchedule:                 "* * * * *",
		},
	}
}

// CurrentExecutionIterator is the iterator of current executions
func CurrentExecutionIterator(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
) pagination.Iterator {
	return CurrentExecutionType.ToIterator()(ctx, pr, params.PageSize)
}

// CurrentExecutionFixerIterator is the iterator of fixer execution
func CurrentExecutionFixerIterator(
	ctx context.Context,
	client blobstore.Client,
	keys store.Keys,
	_ shardscanner.FixShardActivityParams,
) store.ScanOutputIterator {
	return store.NewBlobstoreIterator(ctx, client, keys, CurrentExecutionType.ToBlobstoreEntity())
}

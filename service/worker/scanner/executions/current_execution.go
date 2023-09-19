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
	"fmt"
	"strconv"
	"time"

	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cache"
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

// currentExecutionsScannerHooks provides hooks for current executions scanner.
func currentExecutionsScannerHooks() *shardscanner.ScannerHooks {
	wf, err := shardscanner.NewScannerHooks(currentExecutionScannerManager, currentExecutionScannerIterator, currentExecutionCustomScannerConfig)
	if err != nil {
		return nil
	}

	return wf
}

// currentExecutionScannerManager is the current execution scanner manager
func currentExecutionScannerManager(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
	domainCache cache.DomainCache,
) invariant.Manager {
	var ivs []invariant.Invariant
	collections := ParseCollections(params.ScannerConfig)
	for _, fn := range CurrentExecutionType.ToInvariants(collections, zap.NewNop()) {
		ivs = append(ivs, fn(pr, domainCache))
	}
	return invariant.NewInvariantManager(ivs)
}

// currentExecutionFixerManager provides invariant manager for concrete execution fixer.
func currentExecutionFixerManager(_ context.Context, pr persistence.Retryer, params shardscanner.FixShardActivityParams, domainCache cache.DomainCache) invariant.Manager {
	var ivs []invariant.Invariant
	var collections []invariant.Collection

	if len(params.EnabledInvariants) == 0 {
		// if none, fall back to historical behavior.  this may be removed.
		collections = append(collections, invariant.CollectionHistory, invariant.CollectionMutableState)
	} else {
		// convert to invariants.
		// this may produce an empty list if it all fixers are intentionally disabled.
		for k, v := range params.EnabledInvariants {
			if v == strconv.FormatBool(true) {
				ivc, err := invariant.CollectionString(k)
				if err != nil {
					panic(fmt.Sprintf("invalid collection name: %v", err)) // error includes the name
				}
				collections = append(collections, ivc)
			}
		}
	}

	for _, fn := range CurrentExecutionType.ToInvariants(collections, zap.NewNop()) {
		ivs = append(ivs, fn(pr, domainCache))
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

// currentExecutionCustomScannerConfig resolves dynamic config for current executions scanner.
func currentExecutionCustomScannerConfig(ctx shardscanner.Context) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}

	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerInvariantCollectionHistory)() {
		res[invariant.CollectionHistory.String()] = strconv.FormatBool(true) // TODO: this appears unused, executions/types.go for current does not check this const
	}
	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerInvariantCollectionMutableState)() {
		res[invariant.CollectionMutableState.String()] = strconv.FormatBool(true)
	}

	return res
}

// currentExecutionCustomFixerConfig resolves dynamic config for current executions fixer.
func currentExecutionCustomFixerConfig(ctx shardscanner.FixerContext) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}
	res[invariant.CollectionMutableState.String()] = strconv.FormatBool(
		ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.CurrentExecutionsFixerInvariantCollectionMutableState)(),
	)
	return res
}

// currentExecutionFixerHooks provides hooks for current executions fixer.
func currentExecutionFixerHooks() *shardscanner.FixerHooks {
	h, err := shardscanner.NewFixerHooks(currentExecutionFixerManager, currentExecutionFixerIterator, currentExecutionCustomFixerConfig)
	if err != nil {
		return nil
	}
	return h
}

// CurrentExecutionConfig configures current execution scanner
func CurrentExecutionConfig(dc *dynamicconfig.Collection) *shardscanner.ScannerConfig {
	return &shardscanner.ScannerConfig{
		ScannerWFTypeName: CurrentExecutionsScannerWFTypeName,
		FixerWFTypeName:   CurrentExecutionsFixerWFTypeName,
		DynamicCollection: dc,
		DynamicParams: shardscanner.DynamicParams{
			ScannerEnabled:          dc.GetBoolProperty(dynamicconfig.CurrentExecutionsScannerEnabled),
			FixerEnabled:            dc.GetBoolProperty(dynamicconfig.CurrentExecutionFixerEnabled),
			Concurrency:             dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerConcurrency),
			PageSize:                dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerPersistencePageSize),
			BlobstoreFlushThreshold: dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerBlobstoreFlushThreshold),
			ActivityBatchSize:       dc.GetIntProperty(dynamicconfig.CurrentExecutionsScannerActivityBatchSize),
			AllowDomain:             dc.GetBoolPropertyFilteredByDomain(dynamicconfig.CurrentExecutionFixerDomainAllow),
		},
		ScannerHooks: currentExecutionsScannerHooks,
		FixerHooks:   currentExecutionFixerHooks,
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

// currentExecutionScannerIterator is the iterator of current executions
func currentExecutionScannerIterator(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
) pagination.Iterator {
	return CurrentExecutionType.ToIterator()(ctx, pr, params.PageSize)
}

// currentExecutionFixerIterator is the iterator of fixer execution
func currentExecutionFixerIterator(
	ctx context.Context,
	client blobstore.Client,
	keys store.Keys,
	_ shardscanner.FixShardActivityParams,
) store.ScanOutputIterator {
	return store.NewBlobstoreIterator(ctx, client, keys, CurrentExecutionType.ToBlobstoreEntity())
}

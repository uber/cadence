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
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

const (
	// ConcreteExecutionsScannerWFTypeName defines workflow type name for concrete executions scanner
	ConcreteExecutionsScannerWFTypeName   = "cadence-sys-executions-scanner-workflow"
	concreteExecutionsScannerWFID         = "cadence-sys-executions-scanner"
	concreteExecutionsScannerTaskListName = "cadence-sys-executions-scanner-tasklist-0"

	// ConcreteExecutionsFixerWFTypeName defines workflow type name for concrete executions fixer
	ConcreteExecutionsFixerWFTypeName   = "cadence-sys-executions-fixer-workflow"
	concreteExecutionsFixerTaskListName = "cadence-sys-executions-fixer-tasklist-0"
)

// ConcreteScannerWorkflow starts concrete executions scanner.
func ConcreteScannerWorkflow(ctx workflow.Context, params shardscanner.ScannerWorkflowParams) error {
	wf, err := shardscanner.NewScannerWorkflow(ctx, ConcreteExecutionsScannerWFTypeName, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

// ConcreteFixerWorkflow starts concrete executions fixer.
func ConcreteFixerWorkflow(
	ctx workflow.Context,
	params shardscanner.FixerWorkflowParams,
) error {

	wf, err := shardscanner.NewFixerWorkflow(ctx, ConcreteExecutionsFixerWFTypeName, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

//ConcreteExecutionHooks provides hooks for concrete executions scanner
func ConcreteExecutionHooks() *shardscanner.ScannerHooks {
	h, err := shardscanner.NewScannerHooks(ScannerManager, ScannerIterator)
	if err != nil {
		return nil
	}
	h.SetConfig(ConcreteExecutionConfig)

	return h
}

// ConcreteExecutionFixerHooks provides hooks needed for concrete executions fixer.
func ConcreteExecutionFixerHooks() *shardscanner.FixerHooks {
	h, err := shardscanner.NewFixerHooks(FixerManager, FixerIterator)
	if err != nil {
		return nil
	}
	return h
}

// ScannerManager provides invariant manager for concrete execution scanner
func ScannerManager(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
	_ shardscanner.ScannerConfig,
) invariant.Manager {

	collections := ParseCollections(params.ScannerConfig)

	var ivs []invariant.Invariant
	for _, fn := range ConcreteExecutionType.ToInvariants(collections) {
		ivs = append(ivs, fn(pr))
	}

	return invariant.NewInvariantManager(ivs)
}

// ScannerIterator provides iterator for concrete execution scanner.
func ScannerIterator(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
	_ shardscanner.ScannerConfig,
) pagination.Iterator {
	it := ConcreteExecutionType.ToIterator()
	return it(ctx, pr, params.PageSize)

}

// FixerIterator provides iterator for concrete execution fixer.
func FixerIterator(ctx context.Context, client blobstore.Client, keys store.Keys, _ shardscanner.FixShardActivityParams, _ shardscanner.ScannerConfig) store.ScanOutputIterator {
	return store.NewBlobstoreIterator(ctx, client, keys, ConcreteExecutionType.ToBlobstoreEntity())
}

// FixerManager provides invariant manager for concrete execution fixer.
func FixerManager(_ context.Context, pr persistence.Retryer, _ shardscanner.FixShardActivityParams, _ shardscanner.ScannerConfig) invariant.Manager {
	var ivs []invariant.Invariant
	var collections []invariant.Collection

	collections = append(collections, invariant.CollectionHistory, invariant.CollectionMutableState)

	for _, fn := range ConcreteExecutionType.ToInvariants(collections) {
		ivs = append(ivs, fn(pr))
	}
	return invariant.NewInvariantManager(ivs)
}

// ConcreteExecutionConfig resolves dynamic config for concrete executions scanner.
func ConcreteExecutionConfig(ctx shardscanner.Context) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}

	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.ConcreteExecutionsScannerInvariantCollectionHistory, true)() {
		res[invariant.CollectionHistory.String()] = strconv.FormatBool(true)
	}
	if ctx.Config.DynamicCollection.GetBoolProperty(dynamicconfig.ConcreteExecutionsScannerInvariantCollectionMutableState, true)() {
		res[invariant.CollectionMutableState.String()] = strconv.FormatBool(true)
	}

	return res
}

// ConcreteExecutionScannerConfig configures concrete execution scanner
func ConcreteExecutionScannerConfig(dc *dynamicconfig.Collection) *shardscanner.ScannerConfig {
	return &shardscanner.ScannerConfig{
		ScannerWFTypeName: ConcreteExecutionsScannerWFTypeName,
		FixerWFTypeName:   ConcreteExecutionsFixerWFTypeName,
		DynamicParams: shardscanner.DynamicParams{
			ScannerEnabled:          dc.GetBoolProperty(dynamicconfig.ConcreteExecutionsScannerEnabled, false),
			FixerEnabled:            dc.GetBoolProperty(dynamicconfig.ConcreteExecutionFixerEnabled, false),
			Concurrency:             dc.GetIntProperty(dynamicconfig.ConcreteExecutionsScannerConcurrency, 25),
			PageSize:                dc.GetIntProperty(dynamicconfig.ConcreteExecutionsScannerPersistencePageSize, 1000),
			BlobstoreFlushThreshold: dc.GetIntProperty(dynamicconfig.ConcreteExecutionsScannerBlobstoreFlushThreshold, 100),
			ActivityBatchSize:       dc.GetIntProperty(dynamicconfig.ConcreteExecutionsScannerActivityBatchSize, 25),
			AllowDomain:             dc.GetBoolPropertyFilteredByDomain(dynamicconfig.ConcreteExecutionFixerDomainAllow, false),
		},
		DynamicCollection: dc,
		ScannerHooks:      ConcreteExecutionHooks,
		FixerHooks:        ConcreteExecutionFixerHooks,
		FixerTLName:       concreteExecutionsFixerTaskListName,
		StartWorkflowOptions: cclient.StartWorkflowOptions{
			ID:                           concreteExecutionsScannerWFID,
			TaskList:                     concreteExecutionsScannerTaskListName,
			ExecutionStartToCloseTimeout: 20 * 365 * 24 * time.Hour,
			WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyAllowDuplicate,
			CronSchedule:                 "* * * * *",
		},
	}
}

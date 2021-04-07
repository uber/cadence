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

package timers

import (
	"context"
	"strconv"
	"time"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/reconciliation/fetcher"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
)

const (
	// ScannerWFTypeName defines workflow type name for concrete executions scanner
	ScannerWFTypeName   = "cadence-sys-timers-scanner-workflow"
	wfid                = "cadence-sys-timers-scanner"
	scannerTaskListName = "cadence-sys-timers-scanner-tasklist-0"

	// FixerWFTypeName defines workflow type name for timers fixer
	FixerWFTypeName   = "cadence-sys-timers-fixer-workflow"
	fixerTaskListName = "cadence-sys-timers-fixer-tasklist-0"
	fixerwfid         = "cadence-sys-timers-fixer"
	periodStartKey    = "period_start"
	periodEndKey      = "period_end"

	// period is defined in hours
	_defaultPeriodStart = 24
	_defaultPeriodEnd   = 3
)

// ScannerWorkflow starts timers scanner.
func ScannerWorkflow(
	ctx workflow.Context,
	params shardscanner.ScannerWorkflowParams,
) error {
	wf, err := shardscanner.NewScannerWorkflow(ctx, ScannerWFTypeName, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

// FixerWorkflow starts timers executions fixer.
func FixerWorkflow(
	ctx workflow.Context,
	params shardscanner.FixerWorkflowParams,
) error {
	wf, err := shardscanner.NewFixerWorkflow(ctx, FixerWFTypeName, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

// ScannerHooks provides hooks for timers scanner.
func ScannerHooks() *shardscanner.ScannerHooks {
	h, err := shardscanner.NewScannerHooks(Manager, Iterator)
	if err != nil {
		return nil
	}
	h.SetConfig(Config)

	return h
}

// FixerHooks provides hooks needed for timers fixer.
func FixerHooks() *shardscanner.FixerHooks {
	h, err := shardscanner.NewFixerHooks(FixerManager, FixerIterator)
	if err != nil {
		return nil
	}
	return h
}

// Manager provides invariant manager for timers scanner.
func Manager(
	_ context.Context,
	pr persistence.Retryer,
	_ shardscanner.ScanShardActivityParams,
) invariant.Manager {
	return invariant.NewInvariantManager(getInvariants(pr))
}

// Iterator provides iterator for timers scanner.
func Iterator(
	ctx context.Context,
	pr persistence.Retryer,
	params shardscanner.ScanShardActivityParams,
) pagination.Iterator {
	start, err := strconv.Atoi(params.ScannerConfig[periodStartKey])
	if err != nil {
		return nil
	}
	end, err := strconv.Atoi(params.ScannerConfig[periodEndKey])
	if err != nil {
		return nil
	}

	startAt := time.Now().Add(time.Hour * time.Duration(start))
	endAt := startAt.Add(time.Hour * time.Duration(end))

	return fetcher.TimerIterator(ctx, pr, startAt, endAt, params.PageSize)

}

// FixerIterator provides iterator for concrete execution fixer.
func FixerIterator(
	ctx context.Context,
	client blobstore.Client,
	keys store.Keys,
	_ shardscanner.FixShardActivityParams,
) store.ScanOutputIterator {
	return store.NewBlobstoreIterator(ctx, client, keys, &entity.Timer{})
}

// FixerManager provides invariant manager for concrete execution fixer.
func FixerManager(
	_ context.Context,
	pr persistence.Retryer,
	_ shardscanner.FixShardActivityParams,
) invariant.Manager {
	return invariant.NewInvariantManager(getInvariants(pr))
}

// Config resolves dynamic config for timers scanner.
func Config(ctx shardscanner.Context) shardscanner.CustomScannerConfig {
	res := shardscanner.CustomScannerConfig{}
	res[periodStartKey] = strconv.Itoa(ctx.Config.DynamicCollection.GetIntProperty(dynamicconfig.TimersScannerPeriodStart, _defaultPeriodStart)())
	res[periodEndKey] = strconv.Itoa(ctx.Config.DynamicCollection.GetIntProperty(dynamicconfig.TimersScannerPeriodEnd, _defaultPeriodEnd)())
	return res
}

// ScannerConfig configures timers scanner
func ScannerConfig(dc *dynamicconfig.Collection) *shardscanner.ScannerConfig {
	return &shardscanner.ScannerConfig{
		ScannerWFTypeName: ScannerWFTypeName,
		FixerWFTypeName:   FixerWFTypeName,
		DynamicParams: shardscanner.DynamicParams{
			ScannerEnabled:          dc.GetBoolProperty(dynamicconfig.TimersScannerEnabled, false),
			FixerEnabled:            dc.GetBoolProperty(dynamicconfig.TimersFixerEnabled, false),
			Concurrency:             dc.GetIntProperty(dynamicconfig.TimersScannerConcurrency, 5),
			PageSize:                dc.GetIntProperty(dynamicconfig.TimersScannerPersistencePageSize, 1000),
			BlobstoreFlushThreshold: dc.GetIntProperty(dynamicconfig.TimersScannerBlobstoreFlushThreshold, 100),
			ActivityBatchSize:       dc.GetIntProperty(dynamicconfig.TimersScannerActivityBatchSize, 25),
			AllowDomain:             dc.GetBoolPropertyFilteredByDomain(dynamicconfig.TimersFixerDomainAllow, false),
		},
		DynamicCollection: dc,
		ScannerHooks:      ScannerHooks,
		FixerHooks:        FixerHooks,

		StartWorkflowOptions: client.StartWorkflowOptions{
			ID:                           wfid,
			TaskList:                     scannerTaskListName,
			ExecutionStartToCloseTimeout: 20 * 365 * 24 * time.Hour,
			WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
			CronSchedule:                 "* * * * *",
		},
		StartFixerOptions: client.StartWorkflowOptions{
			ID:                           fixerwfid,
			TaskList:                     fixerTaskListName,
			ExecutionStartToCloseTimeout: 20 * 365 * 24 * time.Hour,
			WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
			CronSchedule:                 "* * * * *",
		},
	}
}

func getInvariants(pr persistence.Retryer) []invariant.Invariant {
	return []invariant.Invariant{
		invariant.NewTimerInvalid(pr),
	}
}

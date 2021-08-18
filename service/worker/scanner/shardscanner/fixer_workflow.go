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

package shardscanner

import (
	"context"
	"errors"

	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
)

const (
	fixShardReportChan = "fixShardReportChan"
)

var (
	errQueryNotReady = errors.New("query is not yet ready to be handled, please try again shortly")
)

// FixerManagerCB is a function which returns invariant manager for fixer.
type FixerManagerCB func(
	context.Context,
	persistence.Retryer,
	FixShardActivityParams,
) invariant.Manager

// FixerIteratorCB is a function which returns ScanOutputIterator for fixer.
type FixerIteratorCB func(
	context.Context,
	blobstore.Client,
	store.Keys,
	FixShardActivityParams,
) store.ScanOutputIterator

// FixerHooks holds callback functions for shard scanner workflow implementation.
type FixerHooks struct {
	InvariantManager FixerManagerCB
	Iterator         FixerIteratorCB
}

// NewFixerHooks returns initialized callbacks for shard scanner workflow implementation.
func NewFixerHooks(
	manager FixerManagerCB,
	iterator FixerIteratorCB,
) (*FixerHooks, error) {
	if manager == nil || iterator == nil {
		return nil, errors.New("manager or iterator not provided")
	}
	return &FixerHooks{
		InvariantManager: manager,
		Iterator:         iterator,
	}, nil
}

// FixerWorkflow is the workflow that fixes all entities from a scan output.
type FixerWorkflow struct {
	Aggregator *ShardFixResultAggregator
	Params     FixerWorkflowParams
	Keys       *FixerCorruptedKeysActivityResult
}

// NewFixerWorkflow returns a new instance of fixer workflow
func NewFixerWorkflow(
	ctx workflow.Context,
	name string,
	params FixerWorkflowParams,
) (*FixerWorkflow, error) {

	if len(name) < 1 {
		return nil, errors.New("workflow name is not provided")
	}

	wf := FixerWorkflow{
		Params: params,
	}

	corruptKeys, err := GetCorruptedKeys(ctx, wf.Params)
	if err != nil {
		return nil, err
	}
	if corruptKeys.CorruptedKeys == nil {
		return nil, errors.New("corrupted keys not found")
	}

	wf.Keys = corruptKeys
	wf.Aggregator = NewShardFixResultAggregator(corruptKeys.CorruptedKeys, *corruptKeys.MinShard, *corruptKeys.MaxShard)

	for name, fn := range setHandlers(wf.Aggregator) {
		if err := workflow.SetQueryHandler(ctx, name, fn); err != nil {
			return nil, err
		}
	}

	return &wf, nil
}

// Start starts a shard fixer workflow.
func (fx *FixerWorkflow) Start(ctx workflow.Context) error {

	resolvedConfig := resolveFixerConfig(fx.Params.FixerWorkflowConfigOverwrites)
	shardReportChan := workflow.GetSignalChannel(ctx, fixShardReportChan)

	for i := 0; i < resolvedConfig.Concurrency; i++ {
		idx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			batches := getCorruptedKeysBatches(resolvedConfig.ActivityBatchSize, resolvedConfig.Concurrency, fx.Keys.CorruptedKeys, idx)
			for _, batch := range batches {
				activityCtx := getLongActivityContext(ctx)
				var reports []FixReport
				if err := workflow.ExecuteActivity(activityCtx, ActivityFixShard, FixShardActivityParams{
					CorruptedKeysEntries:        batch,
					ResolvedFixerWorkflowConfig: resolvedConfig,
				}).Get(ctx, &reports); err != nil {
					errStr := err.Error()
					shardReportChan.Send(ctx, FixReportError{
						Reports:  nil,
						ErrorStr: &errStr,
					})
					return
				}
				shardReportChan.Send(ctx, FixReportError{
					Reports:  reports,
					ErrorStr: nil,
				})
			}
		})
	}

	for i := 0; i < len(fx.Keys.CorruptedKeys); {
		var reportErr FixReportError
		shardReportChan.Receive(ctx, &reportErr)
		if reportErr.ErrorStr != nil {
			return errors.New(*reportErr.ErrorStr)
		}
		for _, report := range reportErr.Reports {
			fx.Aggregator.AddReport(report)
			i++
		}
	}
	return nil
}

func resolveFixerConfig(overwrites FixerWorkflowConfigOverwrites) ResolvedFixerWorkflowConfig {
	resolvedConfig := ResolvedFixerWorkflowConfig{
		Concurrency:             25,
		BlobstoreFlushThreshold: 1000,
		ActivityBatchSize:       200,
	}
	if overwrites.Concurrency != nil {
		resolvedConfig.Concurrency = *overwrites.Concurrency
	}
	if overwrites.BlobstoreFlushThreshold != nil {
		resolvedConfig.BlobstoreFlushThreshold = *overwrites.BlobstoreFlushThreshold
	}
	if overwrites.ActivityBatchSize != nil {
		resolvedConfig.ActivityBatchSize = *overwrites.ActivityBatchSize
	}
	return resolvedConfig
}

func setHandlers(aggregator *ShardFixResultAggregator) map[string]interface{} {
	return map[string]interface{}{
		ShardReportQuery: func(shardID int) (*FixReport, error) {
			if aggregator == nil {
				return nil, errQueryNotReady
			}
			return aggregator.GetReport(shardID)
		},
		ShardStatusQuery: func(req PaginatedShardQueryRequest) (*ShardStatusQueryResult, error) {
			if aggregator == nil {
				return nil, errQueryNotReady
			}
			return aggregator.GetStatusResult(req)
		},
		ShardStatusSummaryQuery: func() (ShardStatusSummaryResult, error) {
			if aggregator == nil {
				return nil, errQueryNotReady
			}
			return aggregator.GetStatusSummary(), nil
		},
		AggregateReportQuery: func() (AggregateFixReportResult, error) {
			if aggregator == nil {
				return AggregateFixReportResult{}, errQueryNotReady
			}
			return aggregator.GetAggregation(), nil
		},
		DomainReportQuery: func(req DomainReportQueryRequest) (*DomainFixReportQueryResult, error) {
			if aggregator == nil {
				return nil, errQueryNotReady
			}
			return aggregator.GetDomainStatus(req)
		},
	}
}

func getCorruptedKeysBatches(
	batchSize int,
	concurrency int,
	corruptedKeys []CorruptedKeysEntry,
	workerIdx int,
) [][]CorruptedKeysEntry {
	batchIndices := getBatchIndices(batchSize, concurrency, len(corruptedKeys), workerIdx)
	var result [][]CorruptedKeysEntry
	for _, batch := range batchIndices {
		var curr []CorruptedKeysEntry
		for _, i := range batch {
			curr = append(curr, corruptedKeys[i])
		}
		result = append(result, curr)
	}
	return result
}

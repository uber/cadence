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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/store"
)

type fixerWorkflowSuite struct {
	suite.Suite
}

func TestFixerWorkflowSuite(t *testing.T) {
	suite.Run(t, new(fixerWorkflowSuite))
}

func (s *fixerWorkflowSuite) TestResolveFixerConfig() {
	result := resolveFixerConfig(FixerWorkflowConfigOverwrites{
		Concurrency: common.IntPtr(1000),
	})
	s.Equal(ResolvedFixerWorkflowConfig{
		Concurrency:             1000,
		BlobstoreFlushThreshold: 1000,
		ActivityBatchSize:       200,
	}, result)
}

func (s *fixerWorkflowSuite) TestGetCorruptedKeysBatches() {
	var keys []CorruptedKeysEntry
	for i := 5; i < 50; i += 2 {
		keys = append(keys, CorruptedKeysEntry{
			ShardID: i,
		})
	}
	batches := getCorruptedKeysBatches(5, 3, keys, 1)
	s.Equal([][]CorruptedKeysEntry{
		{
			{ShardID: 7},
			{ShardID: 13},
			{ShardID: 19},
			{ShardID: 25},
			{ShardID: 31},
		},
		{
			{ShardID: 37},
			{ShardID: 43},
			{ShardID: 49},
		},
	}, batches)
}

func (s *fixerWorkflowSuite) TestNewFixerHooks() {
	testCases := []struct {
		name     string
		manager  FixerManagerCB
		iterator FixerIteratorCB
		wantErr  bool
	}{
		{
			name:     "both arguments are nil",
			manager:  nil,
			iterator: nil,
			wantErr:  true,
		},
		{
			name: "invariant is nil",
			manager: func(
				ctx context.Context,
				retryer persistence.Retryer,
				params FixShardActivityParams,
				cache cache.DomainCache,
			) invariant.Manager {
				return nil
			},
			iterator: nil,
			wantErr:  true,
		},
		{
			name: "both provided",
			manager: func(
				ctx context.Context,
				retryer persistence.Retryer,
				params FixShardActivityParams,
				cache cache.DomainCache,
			) invariant.Manager {
				return nil
			},
			iterator: func(
				ctx context.Context,
				client blobstore.Client,
				keys store.Keys,
				params FixShardActivityParams,
			) store.ScanOutputIterator {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := NewFixerHooks(tc.manager, tc.iterator, func(fixer FixerContext) CustomScannerConfig {
				return nil // no config overrides
			})
			if tc.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

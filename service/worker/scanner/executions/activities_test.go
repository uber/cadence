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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service/dynamicconfig"
)

type activitiesSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	controller   *gomock.Controller
	mockResource *resource.Test
}

func TestActivitiesSuite(t *testing.T) {
	suite.Run(t, new(activitiesSuite))
}

func (s *activitiesSuite) SetupTest() {
	activity.Register(ScannerConfigActivity)
	activity.Register(ScanShardActivity)
	activity.Register(ScannerEmitMetricsActivity)

	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.controller, metrics.Worker)
}

func (s *activitiesSuite) TestScannerConfigActivity() {
	testCases := []struct {
		scannerWorkflowDynamicConfig *ScannerWorkflowDynamicConfig
		params                       ScannerConfigActivityParams
		resolved                     ResolvedScannerWorkflowConfig
	}{
		{
			scannerWorkflowDynamicConfig: &ScannerWorkflowDynamicConfig{
				Enabled:                 dynamicconfig.GetBoolPropertyFn(true),
				Concurrency:             dynamicconfig.GetIntPropertyFn(10),
				ExecutionsPageSize:      dynamicconfig.GetIntPropertyFn(100),
				BlobstoreFlushThreshold: dynamicconfig.GetIntPropertyFn(1000),
				DynamicConfigInvariantCollections: DynamicConfigInvariantCollections{
					InvariantCollectionMutableState: dynamicconfig.GetBoolPropertyFn(true),
					InvariantCollectionHistory:      dynamicconfig.GetBoolPropertyFn(false),
				},
			},
			params: ScannerConfigActivityParams{
				Overwrites: ScannerWorkflowConfigOverwrites{},
			},
			resolved: ResolvedScannerWorkflowConfig{
				Enabled:                 true,
				Concurrency:             10,
				ExecutionsPageSize:      100,
				BlobstoreFlushThreshold: 1000,
				InvariantCollections: InvariantCollections{
					InvariantCollectionHistory:      false,
					InvariantCollectionMutableState: true,
				},
			},
		},
		{
			scannerWorkflowDynamicConfig: &ScannerWorkflowDynamicConfig{
				Enabled:                 dynamicconfig.GetBoolPropertyFn(true),
				Concurrency:             dynamicconfig.GetIntPropertyFn(10),
				ExecutionsPageSize:      dynamicconfig.GetIntPropertyFn(100),
				BlobstoreFlushThreshold: dynamicconfig.GetIntPropertyFn(1000),
				DynamicConfigInvariantCollections: DynamicConfigInvariantCollections{
					InvariantCollectionMutableState: dynamicconfig.GetBoolPropertyFn(true),
					InvariantCollectionHistory:      dynamicconfig.GetBoolPropertyFn(false),
				},
			},
			params: ScannerConfigActivityParams{
				Overwrites: ScannerWorkflowConfigOverwrites{
					Enabled: common.BoolPtr(false),
					InvariantCollections: &InvariantCollections{
						InvariantCollectionMutableState: false,
						InvariantCollectionHistory:      true,
					},
				},
			},
			resolved: ResolvedScannerWorkflowConfig{
				Enabled:                 false,
				Concurrency:             10,
				ExecutionsPageSize:      100,
				BlobstoreFlushThreshold: 1000,
				InvariantCollections: InvariantCollections{
					InvariantCollectionHistory:      true,
					InvariantCollectionMutableState: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		env := s.NewTestActivityEnvironment()
		env.SetWorkerOptions(worker.Options{
			BackgroundActivityContext: context.WithValue(context.Background(), ScannerContextKey, ScannerContext{
				ScannerWorkflowDynamicConfig: tc.scannerWorkflowDynamicConfig,
			}),
		})
		resolvedValue, err := env.ExecuteActivity(ScannerConfigActivity, tc.params)
		s.NoError(err)
		var resolved ResolvedScannerWorkflowConfig
		s.NoError(resolvedValue.Get(&resolved))
		s.Equal(tc.resolved, resolved)
	}
}

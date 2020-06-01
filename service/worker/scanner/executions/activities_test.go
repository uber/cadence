package executions

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"testing"
)

type activitiesSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	controller *gomock.Controller
	mockResource      *resource.Test
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
	testCases := []struct{
		scannerWorkflowDynamicConfig *ScannerWorkflowDynamicConfig
		params ScannerConfigActivityParams
		resolved ResolvedScannerWorkflowConfig
	}{
		{
			scannerWorkflowDynamicConfig: &ScannerWorkflowDynamicConfig{
				Enabled: dynamicconfig.GetBoolPropertyFn(true),
				Concurrency: dynamicconfig.GetIntPropertyFn(10),
				ExecutionsPageSize: dynamicconfig.GetIntPropertyFn(100),
				BlobstoreFlushThreshold: dynamicconfig.GetIntPropertyFn(1000),
				DynamicConfigInvariantCollections: DynamicConfigInvariantCollections{
					InvariantCollectionMutableState: dynamicconfig.GetBoolPropertyFn(true),
					InvariantCollectionHistory: dynamicconfig.GetBoolPropertyFn(false),
				},
			},
			params: ScannerConfigActivityParams{
				Overwrites: ScannerWorkflowConfigOverwrites{},
			},
			resolved: ResolvedScannerWorkflowConfig{
				Enabled: true,
				Concurrency: 10,
				ExecutionsPageSize: 100,
				BlobstoreFlushThreshold: 1000,
				InvariantCollections: InvariantCollections{
					InvariantCollectionHistory: false,
					InvariantCollectionMutableState: true,
				},
			},
		},
		{
			scannerWorkflowDynamicConfig: &ScannerWorkflowDynamicConfig{
				Enabled: dynamicconfig.GetBoolPropertyFn(true),
				Concurrency: dynamicconfig.GetIntPropertyFn(10),
				ExecutionsPageSize: dynamicconfig.GetIntPropertyFn(100),
				BlobstoreFlushThreshold: dynamicconfig.GetIntPropertyFn(1000),
				DynamicConfigInvariantCollections: DynamicConfigInvariantCollections{
					InvariantCollectionMutableState: dynamicconfig.GetBoolPropertyFn(true),
					InvariantCollectionHistory: dynamicconfig.GetBoolPropertyFn(false),
				},
			},
			params: ScannerConfigActivityParams{
				Overwrites: ScannerWorkflowConfigOverwrites{
					Enabled: common.BoolPtr(false),
					InvariantCollections: &InvariantCollections{
						InvariantCollectionMutableState: false,
						InvariantCollectionHistory: true,
					},
				},
			},
			resolved: ResolvedScannerWorkflowConfig{
				Enabled: false,
				Concurrency: 10,
				ExecutionsPageSize: 100,
				BlobstoreFlushThreshold: 1000,
				InvariantCollections: InvariantCollections{
					InvariantCollectionHistory: true,
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
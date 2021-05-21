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

package decision

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
)

type (
	attrValidatorSuite struct {
		suite.Suite
		*require.Assertions

		controller      *gomock.Controller
		mockDomainCache *cache.MockDomainCache

		validator *attrValidator

		testDomainID       string
		testTargetDomainID string

		testActivityMaxScheduleToStartTimeoutForRetryInSeconds int32
	}
)

func TestAttrValidatorSuite(t *testing.T) {
	s := new(attrValidatorSuite)
	suite.Run(t, s)
}

func (s *attrValidatorSuite) SetupSuite() {
	s.testDomainID = "test domain ID"
	s.testTargetDomainID = "test target domain ID"
	s.testActivityMaxScheduleToStartTimeoutForRetryInSeconds = 1800
}

func (s *attrValidatorSuite) TearDownSuite() {
}

func (s *attrValidatorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)
	config := &config.Config{
		MaxIDLengthWarnLimit:              dynamicconfig.GetIntPropertyFn(128),
		DomainNameMaxLength:               dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		IdentityMaxLength:                 dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		WorkflowIDMaxLength:               dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		SignalNameMaxLength:               dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		WorkflowTypeMaxLength:             dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		RequestIDMaxLength:                dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		TaskListNameMaxLength:             dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		ActivityIDMaxLength:               dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		ActivityTypeMaxLength:             dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		MarkerNameMaxLength:               dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		TimerIDMaxLength:                  dynamicconfig.GetIntPropertyFilteredByDomain(1000),
		ValidSearchAttributes:             dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
		SearchAttributesNumberOfKeysLimit: dynamicconfig.GetIntPropertyFilteredByDomain(100),
		SearchAttributesSizeOfValueLimit:  dynamicconfig.GetIntPropertyFilteredByDomain(2 * 1024),
		SearchAttributesTotalSizeLimit:    dynamicconfig.GetIntPropertyFilteredByDomain(40 * 1024),
		ActivityMaxScheduleToStartTimeoutForRetry: dynamicconfig.GetDurationPropertyFnFilteredByDomain(
			time.Duration(s.testActivityMaxScheduleToStartTimeoutForRetryInSeconds) * time.Second,
		),
	}
	s.validator = newAttrValidator(
		s.mockDomainCache,
		metrics.NewNoopMetricsClient(),
		config,
		log.NewNoop(),
	)
}

func (s *attrValidatorSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *attrValidatorSuite) TestValidateSignalExternalWorkflowExecutionAttributes() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).AnyTimes()

	var attributes *types.SignalExternalWorkflowExecutionDecisionAttributes

	err := s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.EqualError(err, "BadRequestError{Message: SignalExternalWorkflowExecutionDecisionAttributes is not set on decision.}")

	attributes = &types.SignalExternalWorkflowExecutionDecisionAttributes{}
	err = s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.EqualError(err, "BadRequestError{Message: Execution is nil on decision.}")

	attributes.Execution = &types.WorkflowExecution{}
	attributes.Execution.WorkflowID = "workflow-id"
	err = s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.EqualError(err, "BadRequestError{Message: SignalName is not set on decision.}")

	attributes.Execution.RunID = "run-id"
	err = s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.EqualError(err, "BadRequestError{Message: Invalid RunId set on decision.}")
	attributes.Execution.RunID = constants.TestRunID

	attributes.SignalName = "my signal name"
	err = s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.NoError(err)

	attributes.Input = []byte("test input")
	err = s.validator.validateSignalExternalWorkflowExecutionAttributes(s.testDomainID, s.testTargetDomainID, attributes, metrics.HistoryRespondDecisionTaskCompletedScope)
	s.NoError(err)
}

func (s *attrValidatorSuite) TestValidateUpsertWorkflowSearchAttributes() {
	domainName := "testDomain"
	var attributes *types.UpsertWorkflowSearchAttributesDecisionAttributes

	err := s.validator.validateUpsertWorkflowSearchAttributes(domainName, attributes)
	s.EqualError(err, "BadRequestError{Message: UpsertWorkflowSearchAttributesDecisionAttributes is not set on decision.}")

	attributes = &types.UpsertWorkflowSearchAttributesDecisionAttributes{}
	err = s.validator.validateUpsertWorkflowSearchAttributes(domainName, attributes)
	s.EqualError(err, "BadRequestError{Message: SearchAttributes is not set on decision.}")

	attributes.SearchAttributes = &types.SearchAttributes{}
	err = s.validator.validateUpsertWorkflowSearchAttributes(domainName, attributes)
	s.EqualError(err, "BadRequestError{Message: IndexedFields is empty on decision.}")

	attributes.SearchAttributes.IndexedFields = map[string][]byte{"CustomKeywordField": []byte(`"bytes"`)}
	err = s.validator.validateUpsertWorkflowSearchAttributes(domainName, attributes)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_LocalToLocal() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_LocalToEffectiveLocal_SameCluster() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_LocalToEffectiveLocal_DiffCluster() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestAlternativeClusterName}},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_LocalToGlobal() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_EffectiveLocalToLocal_SameCluster() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_EffectiveLocalToLocal_DiffCluster() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestAlternativeClusterName}},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_EffectiveLocalToEffectiveLocal_SameCluster() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}},
		},
		5678,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_EffectiveLocalToEffectiveLocal_DiffCluster() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestAlternativeClusterName}},
		},
		5678,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_EffectiveLocalToGlobal() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
			},
		},
		5678,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_GlobalToLocal() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_GlobalToEffectiveLocal() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		5678,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
			},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_GlobalToGlobal_DiffDomain() {
	domainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestAlternativeClusterName},
				{ClusterName: cluster.TestCurrentClusterName},
			},
		},
		1234,
		nil,
	)
	targetDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		nil,
	)

	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateCrossDomainCall(s.testDomainID, s.testTargetDomainID)
	s.IsType(&types.BadRequestError{}, err)
}

func (s *attrValidatorSuite) TestValidateCrossDomainCall_GlobalToGlobal_SameDomain() {
	targetDomainID := s.testDomainID

	err := s.validator.validateCrossDomainCall(s.testDomainID, targetDomainID)
	s.Nil(err)
}

func (s *attrValidatorSuite) TestValidateTaskListName() {
	taskList := func(name string) *types.TaskList {
		kind := types.TaskListKindNormal
		return &types.TaskList{Name: name, Kind: &kind}
	}

	testCases := []struct {
		defaultVal  string
		input       *types.TaskList
		output      *types.TaskList
		isOutputErr bool
	}{
		{"tl-1", nil, &types.TaskList{Name: "tl-1"}, false},
		{"", taskList("tl-1"), taskList("tl-1"), false},
		{"tl-1", taskList("tl-1"), taskList("tl-1"), false},
		{"", taskList("/tl-1"), taskList("/tl-1"), false},
		{"", taskList("/__cadence_sys"), taskList("/__cadence_sys"), false},
		{"", nil, &types.TaskList{}, true},
		{"", taskList(""), taskList(""), true},
		{"", taskList(common.ReservedTaskListPrefix), taskList(common.ReservedTaskListPrefix), true},
		{"tl-1", taskList(common.ReservedTaskListPrefix), taskList(common.ReservedTaskListPrefix), true},
		{"", taskList(common.ReservedTaskListPrefix + "tl-1"), taskList(common.ReservedTaskListPrefix + "tl-1"), true},
		{"tl-1", taskList(common.ReservedTaskListPrefix + "tl-1"), taskList(common.ReservedTaskListPrefix + "tl-1"), true},
	}

	for _, tc := range testCases {
		key := tc.defaultVal + "#"
		if tc.input != nil {
			key += tc.input.GetName()
		} else {
			key += "nil"
		}
		s.Run(key, func() {
			output, err := s.validator.validatedTaskList(tc.input, tc.defaultVal, metrics.HistoryRespondDecisionTaskCompletedScope, "domain_name")
			if tc.isOutputErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
			s.EqualValues(tc.output, output)
		})
	}
}

func (s *attrValidatorSuite) TestValidateActivityScheduleAttributes_NoRetryPolicy() {
	wfTimeout := int32(5)
	attributes := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID: "some random activityID",
		ActivityType: &types.ActivityType{
			Name: "some random activity type",
		},
		Domain: s.testDomainID,
		TaskList: &types.TaskList{
			Name: "some random task list",
		},
		Input:                         []byte{1, 2, 3},
		ScheduleToCloseTimeoutSeconds: nil, // not set
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(3),
		StartToCloseTimeoutSeconds:    common.Int32Ptr(3),  // ScheduleToStart + StartToClose > wfTimeout
		HeartbeatTimeoutSeconds:       common.Int32Ptr(10), // larger then wfTimeout
	}

	expectedAttributesAfterValidation := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID:                    attributes.ActivityID,
		ActivityType:                  attributes.ActivityType,
		Domain:                        attributes.Domain,
		TaskList:                      attributes.TaskList,
		Input:                         attributes.Input,
		ScheduleToCloseTimeoutSeconds: common.Int32Ptr(wfTimeout),
		ScheduleToStartTimeoutSeconds: attributes.ScheduleToStartTimeoutSeconds,
		StartToCloseTimeoutSeconds:    attributes.StartToCloseTimeoutSeconds,
		HeartbeatTimeoutSeconds:       common.Int32Ptr(wfTimeout),
	}

	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateActivityScheduleAttributes(
		s.testDomainID,
		s.testTargetDomainID,
		attributes,
		wfTimeout,
		metrics.HistoryRespondDecisionTaskCompletedScope,
	)
	s.Nil(err)
	s.Equal(expectedAttributesAfterValidation, attributes)
}

func (s *attrValidatorSuite) TestValidateActivityScheduleAttributes_WithRetryPolicy_ScheduleToStartRetryable() {
	s.mockDomainCache.EXPECT().GetDomainName(s.testDomainID).Return("some random domain name", nil).Times(1)

	wfTimeout := int32(3000)
	attributes := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID: "some random activityID",
		ActivityType: &types.ActivityType{
			Name: "some random activity type",
		},
		Domain: s.testDomainID,
		TaskList: &types.TaskList{
			Name: "some random task list",
		},
		Input:                         []byte{1, 2, 3},
		ScheduleToCloseTimeoutSeconds: nil, // not set
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(3),
		StartToCloseTimeoutSeconds:    common.Int32Ptr(500), // extended ScheduleToStart + StartToClose > wfTimeout
		HeartbeatTimeoutSeconds:       common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          1.1,
			ExpirationIntervalInSeconds: s.testActivityMaxScheduleToStartTimeoutForRetryInSeconds + 1000, // larger than maximumScheduleToStartTimeoutForRetryInSeconds
			NonRetriableErrorReasons:    []string{"non-retryable error"},
		},
	}

	expectedAttributesAfterValidation := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID:                    attributes.ActivityID,
		ActivityType:                  attributes.ActivityType,
		Domain:                        attributes.Domain,
		TaskList:                      attributes.TaskList,
		Input:                         attributes.Input,
		ScheduleToCloseTimeoutSeconds: common.Int32Ptr(attributes.RetryPolicy.ExpirationIntervalInSeconds),
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(s.testActivityMaxScheduleToStartTimeoutForRetryInSeconds),
		StartToCloseTimeoutSeconds:    attributes.StartToCloseTimeoutSeconds,
		HeartbeatTimeoutSeconds:       attributes.HeartbeatTimeoutSeconds,
		RetryPolicy:                   attributes.RetryPolicy,
	}

	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateActivityScheduleAttributes(
		s.testDomainID,
		s.testTargetDomainID,
		attributes,
		wfTimeout,
		metrics.HistoryRespondDecisionTaskCompletedScope,
	)
	s.Nil(err)
	s.Equal(expectedAttributesAfterValidation, attributes)
}

func (s *attrValidatorSuite) TestValidateActivityScheduleAttributes_WithRetryPolicy_ScheduleToStartNonRetryable() {
	wfTimeout := int32(1000)
	attributes := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID: "some random activityID",
		ActivityType: &types.ActivityType{
			Name: "some random activity type",
		},
		Domain: s.testDomainID,
		TaskList: &types.TaskList{
			Name: "some random task list",
		},
		Input:                         []byte{1, 2, 3},
		ScheduleToCloseTimeoutSeconds: nil, // not set
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(3),
		StartToCloseTimeoutSeconds:    common.Int32Ptr(500), // extended ScheduleToStart + StartToClose > wfTimeout
		HeartbeatTimeoutSeconds:       common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          1.1,
			ExpirationIntervalInSeconds: s.testActivityMaxScheduleToStartTimeoutForRetryInSeconds + 1000, // larger than wfTimeout and maximumScheduleToStartTimeoutForRetryInSeconds
			NonRetriableErrorReasons:    []string{"cadenceInternal:Timeout SCHEDULE_TO_START"},
		},
	}

	expectedAttributesAfterValidation := &types.ScheduleActivityTaskDecisionAttributes{
		ActivityID:                    attributes.ActivityID,
		ActivityType:                  attributes.ActivityType,
		Domain:                        attributes.Domain,
		TaskList:                      attributes.TaskList,
		Input:                         attributes.Input,
		ScheduleToCloseTimeoutSeconds: common.Int32Ptr(wfTimeout),
		ScheduleToStartTimeoutSeconds: attributes.ScheduleToStartTimeoutSeconds,
		StartToCloseTimeoutSeconds:    attributes.StartToCloseTimeoutSeconds,
		HeartbeatTimeoutSeconds:       attributes.HeartbeatTimeoutSeconds,
		RetryPolicy:                   attributes.RetryPolicy,
	}

	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	targetDomainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: s.testTargetDomainID},
		nil,
		cluster.TestCurrentClusterName,
		nil,
	)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testDomainID).Return(domainEntry, nil).Times(1)
	s.mockDomainCache.EXPECT().GetDomainByID(s.testTargetDomainID).Return(targetDomainEntry, nil).Times(1)

	err := s.validator.validateActivityScheduleAttributes(
		s.testDomainID,
		s.testTargetDomainID,
		attributes,
		wfTimeout,
		metrics.HistoryRespondDecisionTaskCompletedScope,
	)
	s.Nil(err)
	s.Equal(expectedAttributesAfterValidation, attributes)
}

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

package frontend

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/client"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/domain"
	dc "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
)

const (
	numHistoryShards = 10

	testWorkflowID            = "test-workflow-id"
	testRunID                 = "2c8b555f-1f55-4955-9d1c-b980194555c9"
	testHistoryArchivalURI    = "testScheme://history/URI"
	testVisibilityArchivalURI = "testScheme://visibility/URI"
)

type (
	workflowHandlerSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockResource        *resource.Test
		mockDomainCache     *cache.MockDomainCache
		mockHistoryClient   *history.MockClient
		mockClusterMetadata *cluster.MockMetadata

		mockProducer           *mocks.KafkaProducer
		mockMessagingClient    messaging.Client
		mockMetadataMgr        *mocks.MetadataManager
		mockHistoryV2Mgr       *mocks.HistoryV2Manager
		mockVisibilityMgr      *mocks.VisibilityManager
		mockArchivalMetadata   *archiver.MockArchivalMetadata
		mockArchiverProvider   *provider.MockArchiverProvider
		mockHistoryArchiver    *archiver.HistoryArchiverMock
		mockVisibilityArchiver *archiver.VisibilityArchiverMock
		mockVersionChecker     *client.VersionCheckerMock

		testDomain   string
		testDomainID string
	}
)

func TestWorkflowHandlerSuite(t *testing.T) {
	s := new(workflowHandlerSuite)
	suite.Run(t, s)
}

func (s *workflowHandlerSuite) SetupSuite() {
}

func (s *workflowHandlerSuite) TearDownSuite() {
}

func (s *workflowHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.testDomain = "test-domain"
	s.testDomainID = "e4f90ec0-1313-45be-9877-8aa41f72a45a"

	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.controller, metrics.Frontend)
	s.mockDomainCache = s.mockResource.DomainCache
	s.mockHistoryClient = s.mockResource.HistoryClient
	s.mockClusterMetadata = s.mockResource.ClusterMetadata
	s.mockMetadataMgr = s.mockResource.MetadataMgr
	s.mockHistoryV2Mgr = s.mockResource.HistoryMgr
	s.mockVisibilityMgr = s.mockResource.VisibilityMgr
	s.mockArchivalMetadata = s.mockResource.ArchivalMetadata
	s.mockArchiverProvider = s.mockResource.ArchiverProvider

	s.mockProducer = &mocks.KafkaProducer{}
	s.mockMessagingClient = mocks.NewMockMessagingClient(s.mockProducer, nil)
	s.mockHistoryArchiver = &archiver.HistoryArchiverMock{}
	s.mockVisibilityArchiver = &archiver.VisibilityArchiverMock{}
	s.mockVersionChecker = client.NewMockVersionChecker(s.controller)

	mockMonitor := s.mockResource.MembershipMonitor
	mockMonitor.EXPECT().GetMemberCount(common.FrontendServiceName).Return(5, nil).AnyTimes()
	s.mockVersionChecker.EXPECT().ClientSupported(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

}

func (s *workflowHandlerSuite) TearDownTest() {
	s.controller.Finish()
	s.mockResource.Finish(s.T())
	s.mockProducer.AssertExpectations(s.T())
	s.mockHistoryArchiver.AssertExpectations(s.T())
	s.mockVisibilityArchiver.AssertExpectations(s.T())
}

func (s *workflowHandlerSuite) getWorkflowHandler(config *Config) *WorkflowHandler {
	return NewWorkflowHandler(s.mockResource, config, s.mockProducer, s.mockVersionChecker)
}

func (s *workflowHandlerSuite) TestDisableListVisibilityByFilter() {
	domain := "test-domain"
	domainID := uuid.New()
	config := s.newConfig(dc.NewInMemoryClient())
	config.DisableListVisibilityByFilter = dc.GetBoolPropertyFnFilteredByDomain(true)

	wh := s.getWorkflowHandler(config)

	s.mockDomainCache.EXPECT().GetDomainID(gomock.Any()).Return(domainID, nil).AnyTimes()

	// test list open by wid
	listRequest := &types.ListOpenWorkflowExecutionsRequest{
		Domain: domain,
		StartTimeFilter: &types.StartTimeFilter{
			EarliestTime: common.Int64Ptr(0),
			LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
		ExecutionFilter: &types.WorkflowExecutionFilter{
			WorkflowID: "wid",
		},
	}
	_, err := wh.ListOpenWorkflowExecutions(context.Background(), listRequest)
	s.Error(err)
	s.Equal(errNoPermission, err)

	// test list open by workflow type
	listRequest.ExecutionFilter = nil
	listRequest.TypeFilter = &types.WorkflowTypeFilter{
		Name: "workflow-type",
	}
	_, err = wh.ListOpenWorkflowExecutions(context.Background(), listRequest)
	s.Error(err)
	s.Equal(errNoPermission, err)

	// test list close by wid
	listRequest2 := &types.ListClosedWorkflowExecutionsRequest{
		Domain: domain,
		StartTimeFilter: &types.StartTimeFilter{
			EarliestTime: common.Int64Ptr(0),
			LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
		ExecutionFilter: &types.WorkflowExecutionFilter{
			WorkflowID: "wid",
		},
	}
	_, err = wh.ListClosedWorkflowExecutions(context.Background(), listRequest2)
	s.Error(err)
	s.Equal(errNoPermission, err)

	// test list close by workflow type
	listRequest2.ExecutionFilter = nil
	listRequest2.TypeFilter = &types.WorkflowTypeFilter{
		Name: "workflow-type",
	}
	_, err = wh.ListClosedWorkflowExecutions(context.Background(), listRequest2)
	s.Error(err)
	s.Equal(errNoPermission, err)

	// test list close by workflow status
	listRequest2.TypeFilter = nil
	failedStatus := types.WorkflowExecutionCloseStatusFailed
	listRequest2.StatusFilter = &failedStatus
	_, err = wh.ListClosedWorkflowExecutions(context.Background(), listRequest2)
	s.Error(err)
	s.Equal(errNoPermission, err)
}

func (s *workflowHandlerSuite) TestPollForTask_Failed_ContextTimeoutTooShort() {
	config := s.newConfig(dc.NewInMemoryClient())
	wh := s.getWorkflowHandler(config)

	bgCtx := context.Background()
	_, err := wh.PollForDecisionTask(bgCtx, &types.PollForDecisionTaskRequest{})
	s.Error(err)
	s.Equal(common.ErrContextTimeoutNotSet, err)

	_, err = wh.PollForActivityTask(bgCtx, &types.PollForActivityTaskRequest{})
	s.Error(err)
	s.Equal(common.ErrContextTimeoutNotSet, err)

	shortCtx, cancel := context.WithTimeout(bgCtx, common.MinLongPollTimeout-time.Millisecond)
	defer cancel()

	_, err = wh.PollForDecisionTask(shortCtx, &types.PollForDecisionTaskRequest{})
	s.Error(err)
	s.Equal(common.ErrContextTimeoutTooShort, err)

	_, err = wh.PollForActivityTask(shortCtx, &types.PollForActivityTaskRequest{})
	s.Error(err)
	s.Equal(common.ErrContextTimeoutTooShort, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_RequestIdNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errRequestIDNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_BadDelayStartSeconds() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID:         uuid.New(),
		DelayStartSeconds: common.Int32Ptr(-1),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errInvalidDelayStartSeconds, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_StartRequestNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	_, err := wh.StartWorkflowExecution(context.Background(), nil)
	s.Error(err)
	s.Equal(errRequestNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_DomainNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errDomainNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_WorkflowIdNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain: "test-domain",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errWorkflowIDNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_WorkflowTypeNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errWorkflowTypeNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_TaskListNotSet() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errTaskListNotSet, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_InvalidExecutionStartToCloseTimeout() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(0),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errInvalidExecutionStartToCloseTimeoutSeconds, err)
}

func (s *workflowHandlerSuite) TestStartWorkflowExecution_Failed_InvalidTaskStartToCloseTimeout() {
	config := s.newConfig(dc.NewInMemoryClient())
	config.RPS = dc.GetIntPropertyFn(10)
	wh := s.getWorkflowHandler(config)

	startWorkflowExecutionRequest := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "workflow-type",
		},
		TaskList: &types.TaskList{
			Name: "task-list",
		},
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(0),
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          2,
			MaximumIntervalInSeconds:    2,
			MaximumAttempts:             1,
			ExpirationIntervalInSeconds: 1,
		},
		RequestID: uuid.New(),
	}
	_, err := wh.StartWorkflowExecution(context.Background(), startWorkflowExecutionRequest)
	s.Error(err)
	s.Equal(errInvalidTaskStartToCloseTimeoutSeconds, err)
}

func (s *workflowHandlerSuite) TestRegisterDomain_Failure_MissingDomainDataKey() {
	dynamicClient := dc.NewInMemoryClient()
	dynamicClient.UpdateValue(dc.RequiredDomainDataKeys, map[string]interface{}{"Tier": true})
	cfg := s.newConfig(dynamicClient)
	wh := s.getWorkflowHandler(cfg)

	req := registerDomainRequest(
		types.ArchivalStatusEnabled.Ptr(),
		testHistoryArchivalURI,
		types.ArchivalStatusEnabled.Ptr(),
		testVisibilityArchivalURI,
	)
	err := wh.RegisterDomain(context.Background(), req)
	s.Error(err)
	s.Contains(err.Error(), "domain data error, missing required key")
}

func (s *workflowHandlerSuite) TestRegisterDomain_Failure_InvalidArchivalURI() {
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false)
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName)
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{})
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(errors.New("invalid URI"))
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := registerDomainRequest(
		types.ArchivalStatusEnabled.Ptr(),
		testHistoryArchivalURI,
		types.ArchivalStatusEnabled.Ptr(),
		testVisibilityArchivalURI,
	)
	err := wh.RegisterDomain(context.Background(), req)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestRegisterDomain_Success_EnabledWithNoArchivalURI() {
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", testHistoryArchivalURI))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", testVisibilityArchivalURI))
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{})
	s.mockMetadataMgr.On("CreateDomain", mock.Anything, mock.Anything).Return(&persistence.CreateDomainResponse{
		ID: "test-id",
	}, nil)
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := registerDomainRequest(types.ArchivalStatusEnabled.Ptr(), "", types.ArchivalStatusEnabled.Ptr(), "")
	err := wh.RegisterDomain(context.Background(), req)
	s.NoError(err)
}

func (s *workflowHandlerSuite) TestRegisterDomain_Success_EnabledWithArchivalURI() {
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "invalidURI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "invalidURI"))
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{})
	s.mockMetadataMgr.On("CreateDomain", mock.Anything, mock.Anything).Return(&persistence.CreateDomainResponse{
		ID: "test-id",
	}, nil)
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := registerDomainRequest(
		types.ArchivalStatusEnabled.Ptr(),
		testHistoryArchivalURI,
		types.ArchivalStatusEnabled.Ptr(),
		testVisibilityArchivalURI,
	)
	err := wh.RegisterDomain(context.Background(), req)
	s.NoError(err)
}

func (s *workflowHandlerSuite) TestRegisterDomain_Success_ClusterNotConfiguredForArchival() {
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{})
	s.mockMetadataMgr.On("CreateDomain", mock.Anything, mock.Anything).Return(&persistence.CreateDomainResponse{
		ID: "test-id",
	}, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := registerDomainRequest(
		types.ArchivalStatusEnabled.Ptr(),
		testVisibilityArchivalURI,
		types.ArchivalStatusEnabled.Ptr(),
		"invalidURI",
	)
	err := wh.RegisterDomain(context.Background(), req)
	s.NoError(err)
}

func (s *workflowHandlerSuite) TestRegisterDomain_Success_NotEnabled() {
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{})
	s.mockMetadataMgr.On("CreateDomain", mock.Anything, mock.Anything).Return(&persistence.CreateDomainResponse{
		ID: "test-id",
	}, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := registerDomainRequest(nil, "", nil, "")
	err := wh.RegisterDomain(context.Background(), req)
	s.NoError(err)
}

func (s *workflowHandlerSuite) TestDescribeDomain_Success_ArchivalDisabled() {
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := &types.DescribeDomainRequest{
		Name: common.StringPtr("test-domain"),
	}
	result, err := wh.DescribeDomain(context.Background(), req)

	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal("", result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal("", result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestDescribeDomain_Success_ArchivalEnabled() {
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testHistoryArchivalURI},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testVisibilityArchivalURI},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	req := &types.DescribeDomainRequest{
		Name: common.StringPtr("test-domain"),
	}
	result, err := wh.DescribeDomain(context.Background(), req)

	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal(testHistoryArchivalURI, result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal(testVisibilityArchivalURI, result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestUpdateDomain_Failure_UpdateExistingArchivalURI() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testHistoryArchivalURI},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testVisibilityArchivalURI},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		nil,
		nil,
		common.StringPtr("updated visibility URI"),
		nil,
	)
	_, err := wh.UpdateDomain(context.Background(), updateReq)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestUpdateDomain_Failure_InvalidArchivalURI() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(errors.New("invalid URI"))
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		common.StringPtr("testScheme://invalid/updated/history/URI"),
		types.ArchivalStatusEnabled.Ptr(),
		nil,
		nil,
	)
	_, err := wh.UpdateDomain(context.Background(), updateReq)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestUpdateDomain_Success_ArchivalEnabledToArchivalDisabledWithoutSettingURI() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testHistoryArchivalURI},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testVisibilityArchivalURI},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything, mock.Anything).Return(nil)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		nil,
		types.ArchivalStatusDisabled.Ptr(),
		nil,
		types.ArchivalStatusDisabled.Ptr(),
	)
	result, err := wh.UpdateDomain(context.Background(), updateReq)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal(testHistoryArchivalURI, result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal(testVisibilityArchivalURI, result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestUpdateDomain_Success_ClusterNotConfiguredForArchival() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: "some random history URI"},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: "some random visibility URI"},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(nil, types.ArchivalStatusDisabled.Ptr(), nil, nil)
	result, err := wh.UpdateDomain(context.Background(), updateReq)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal("some random history URI", result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal("some random visibility URI", result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestUpdateDomain_Success_ArchivalEnabledToArchivalDisabledWithSettingBucket() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testHistoryArchivalURI},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testVisibilityArchivalURI},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything, mock.Anything).Return(nil)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		common.StringPtr(testHistoryArchivalURI),
		types.ArchivalStatusDisabled.Ptr(),
		common.StringPtr(testVisibilityArchivalURI),
		types.ArchivalStatusDisabled.Ptr(),
	)
	result, err := wh.UpdateDomain(context.Background(), updateReq)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal(testHistoryArchivalURI, result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusDisabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal(testVisibilityArchivalURI, result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestUpdateDomain_Success_ArchivalEnabledToEnabled() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testHistoryArchivalURI},
		&domain.ArchivalState{Status: types.ArchivalStatusEnabled, URI: testVisibilityArchivalURI},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		common.StringPtr(testHistoryArchivalURI),
		types.ArchivalStatusEnabled.Ptr(),
		common.StringPtr(testVisibilityArchivalURI),
		types.ArchivalStatusEnabled.Ptr(),
	)
	result, err := wh.UpdateDomain(context.Background(), updateReq)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal(testHistoryArchivalURI, result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal(testVisibilityArchivalURI, result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestUpdateDomain_Success_ArchivalNeverEnabledToEnabled() {
	s.mockMetadataMgr.On("GetMetadata", mock.Anything).Return(&persistence.GetMetadataResponse{
		NotificationVersion: int64(0),
	}, nil)
	getDomainResp := persistenceGetDomainResponse(
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
		&domain.ArchivalState{Status: types.ArchivalStatusDisabled, URI: ""},
	)
	s.mockMetadataMgr.On("GetDomain", mock.Anything, mock.Anything).Return(getDomainResp, nil)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything, mock.Anything).Return(nil)
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockArchivalMetadata.On("GetHistoryConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "some random URI"))
	s.mockHistoryArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockVisibilityArchiver.On("ValidateURI", mock.Anything).Return(nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	updateReq := updateRequest(
		common.StringPtr(testHistoryArchivalURI),
		types.ArchivalStatusEnabled.Ptr(),
		common.StringPtr(testVisibilityArchivalURI),
		types.ArchivalStatusEnabled.Ptr(),
	)
	result, err := wh.UpdateDomain(context.Background(), updateReq)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Configuration)
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetHistoryArchivalStatus())
	s.Equal(testHistoryArchivalURI, result.Configuration.GetHistoryArchivalURI())
	s.Equal(types.ArchivalStatusEnabled, result.Configuration.GetVisibilityArchivalStatus())
	s.Equal(testVisibilityArchivalURI, result.Configuration.GetVisibilityArchivalURI())
}

func (s *workflowHandlerSuite) TestHistoryArchived() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	getHistoryRequest := &types.GetWorkflowExecutionHistoryRequest{}
	s.False(wh.historyArchived(context.Background(), getHistoryRequest, "test-domain"))

	getHistoryRequest = &types.GetWorkflowExecutionHistoryRequest{
		Execution: &types.WorkflowExecution{},
	}
	s.False(wh.historyArchived(context.Background(), getHistoryRequest, "test-domain"))

	s.mockHistoryClient.EXPECT().GetMutableState(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	getHistoryRequest = &types.GetWorkflowExecutionHistoryRequest{
		Execution: &types.WorkflowExecution{
			WorkflowID: testWorkflowID,
			RunID:      testRunID,
		},
	}
	s.False(wh.historyArchived(context.Background(), getHistoryRequest, "test-domain"))

	s.mockHistoryClient.EXPECT().GetMutableState(gomock.Any(), gomock.Any()).Return(nil, &types.EntityNotExistsError{Message: "got archival indication error"}).Times(1)
	getHistoryRequest = &types.GetWorkflowExecutionHistoryRequest{
		Execution: &types.WorkflowExecution{
			WorkflowID: testWorkflowID,
			RunID:      testRunID,
		},
	}
	s.True(wh.historyArchived(context.Background(), getHistoryRequest, "test-domain"))

	s.mockHistoryClient.EXPECT().GetMutableState(gomock.Any(), gomock.Any()).Return(nil, errors.New("got non-archival indication error")).Times(1)
	getHistoryRequest = &types.GetWorkflowExecutionHistoryRequest{
		Execution: &types.WorkflowExecution{
			WorkflowID: testWorkflowID,
			RunID:      testRunID,
		},
	}
	s.False(wh.historyArchived(context.Background(), getHistoryRequest, "test-domain"))
}

func (s *workflowHandlerSuite) TestGetArchivedHistory_Failure_DomainCacheEntryError() {
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(nil, errors.New("error getting domain")).Times(1)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.getArchivedHistory(context.Background(), getHistoryRequest(nil), s.testDomainID, metrics.NoopScope(metrics.Frontend))
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestGetArchivedHistory_Failure_ArchivalURIEmpty() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		&persistence.DomainConfig{
			HistoryArchivalStatus:    types.ArchivalStatusDisabled,
			HistoryArchivalURI:       "",
			VisibilityArchivalStatus: types.ArchivalStatusDisabled,
			VisibilityArchivalURI:    "",
		},
		"",
		nil)
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(domainEntry, nil).AnyTimes()

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.getArchivedHistory(context.Background(), getHistoryRequest(nil), s.testDomainID, metrics.NoopScope(metrics.Frontend))
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestGetArchivedHistory_Failure_InvalidURI() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		&persistence.DomainConfig{
			HistoryArchivalStatus:    types.ArchivalStatusEnabled,
			HistoryArchivalURI:       "uri without scheme",
			VisibilityArchivalStatus: types.ArchivalStatusDisabled,
			VisibilityArchivalURI:    "",
		},
		"",
		nil)
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(domainEntry, nil).AnyTimes()

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.getArchivedHistory(context.Background(), getHistoryRequest(nil), s.testDomainID, metrics.NoopScope(metrics.Frontend))
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestGetArchivedHistory_Success_GetFirstPage() {
	domainEntry := cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		&persistence.DomainConfig{
			HistoryArchivalStatus:    types.ArchivalStatusEnabled,
			HistoryArchivalURI:       testHistoryArchivalURI,
			VisibilityArchivalStatus: types.ArchivalStatusDisabled,
			VisibilityArchivalURI:    "",
		},
		"",
		nil)
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(domainEntry, nil).AnyTimes()

	nextPageToken := []byte{'1', '2', '3'}
	historyBatch1 := &types.History{
		Events: []*types.HistoryEvent{
			{EventID: 1},
			{EventID: 2},
		},
	}
	historyBatch2 := &types.History{
		Events: []*types.HistoryEvent{
			{EventID: 3},
			{EventID: 4},
			{EventID: 5},
		},
	}
	history := &types.History{}
	history.Events = append(history.Events, historyBatch1.Events...)
	history.Events = append(history.Events, historyBatch2.Events...)
	s.mockHistoryArchiver.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&archiver.GetHistoryResponse{
		NextPageToken:  nextPageToken,
		HistoryBatches: []*types.History{historyBatch1, historyBatch2},
	}, nil)
	s.mockArchiverProvider.On("GetHistoryArchiver", mock.Anything, mock.Anything).Return(s.mockHistoryArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.getArchivedHistory(context.Background(), getHistoryRequest(nil), s.testDomainID, metrics.NoopScope(metrics.Frontend))
	s.NoError(err)
	s.NotNil(resp)
	s.NotNil(resp.History)
	s.Equal(history, resp.History)
	s.Equal(nextPageToken, resp.NextPageToken)
	s.True(resp.GetArchived())
}

func (s *workflowHandlerSuite) TestGetHistory() {
	domainID := uuid.New()
	firstEventID := int64(100)
	nextEventID := int64(101)
	branchToken := []byte{1}
	we := types.WorkflowExecution{
		WorkflowID: "wid",
		RunID:      "rid",
	}
	shardID := common.WorkflowIDToHistoryShard(we.WorkflowID, numHistoryShards)
	req := &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    firstEventID,
		MaxEventID:    nextEventID,
		PageSize:      0,
		NextPageToken: []byte{},
		ShardID:       common.IntPtr(shardID),
	}
	s.mockHistoryV2Mgr.On("ReadHistoryBranch", mock.Anything, req).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents: []*types.HistoryEvent{
			{
				EventID: int64(100),
			},
		},
		NextPageToken:    []byte{},
		Size:             1,
		LastFirstEventID: nextEventID,
	}, nil).Once()

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	scope := metrics.NoopScope(metrics.Frontend)
	history, token, err := wh.getHistory(context.Background(), scope, domainID, we, firstEventID, nextEventID, 0, []byte{}, nil, branchToken)
	s.NoError(err)
	s.NotNil(history)
	s.Equal([]byte{}, token)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Failure_InvalidRequest() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), &types.ListArchivedWorkflowExecutionsRequest{})
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Failure_ClusterNotConfiguredForArchival() {
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), listArchivedWorkflowExecutionsTestRequest())
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Failure_DomainCacheEntryError() {
	s.mockDomainCache.EXPECT().GetDomain(gomock.Any()).Return(nil, errors.New("error getting domain"))
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), listArchivedWorkflowExecutionsTestRequest())
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Failure_DomainNotConfiguredForArchival() {
	s.mockDomainCache.EXPECT().GetDomain(gomock.Any()).Return(cache.NewLocalDomainCacheEntryForTest(
		nil,
		&persistence.DomainConfig{
			VisibilityArchivalStatus: types.ArchivalStatusDisabled,
		},
		"",
		nil,
	), nil)
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), listArchivedWorkflowExecutionsTestRequest())
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Failure_InvalidURI() {
	s.mockDomainCache.EXPECT().GetDomain(gomock.Any()).Return(cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		&persistence.DomainConfig{
			VisibilityArchivalStatus: types.ArchivalStatusDisabled,
			VisibilityArchivalURI:    "uri without scheme",
		},
		"",
		nil,
	), nil)
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), listArchivedWorkflowExecutionsTestRequest())
	s.Nil(resp)
	s.Error(err)
}

func (s *workflowHandlerSuite) TestListArchivedVisibility_Success() {
	s.mockDomainCache.EXPECT().GetDomain(gomock.Any()).Return(cache.NewLocalDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		&persistence.DomainConfig{
			VisibilityArchivalStatus: types.ArchivalStatusEnabled,
			VisibilityArchivalURI:    testVisibilityArchivalURI,
		},
		"",
		nil,
	), nil).AnyTimes()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))
	s.mockVisibilityArchiver.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(&archiver.QueryVisibilityResponse{}, nil)
	s.mockArchiverProvider.On("GetVisibilityArchiver", mock.Anything, mock.Anything).Return(s.mockVisibilityArchiver, nil)

	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	resp, err := wh.ListArchivedWorkflowExecutions(context.Background(), listArchivedWorkflowExecutionsTestRequest())
	s.NotNil(resp)
	s.NoError(err)
}

func (s *workflowHandlerSuite) TestGetSearchAttributes() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	ctx := context.Background()
	resp, err := wh.GetSearchAttributes(ctx)
	s.NoError(err)
	s.NotNil(resp)
}

func (s *workflowHandlerSuite) TestGetWorkflowExecutionHistory__Success__RawHistoryEnabledTransientDecisionEmitted() {
	var nextEventID int64 = 5
	s.getWorkflowExecutionHistory(5, &types.TransientDecisionInfo{
		StartedEvent:   &types.HistoryEvent{EventID: nextEventID + 1},
		ScheduledEvent: &types.HistoryEvent{EventID: nextEventID},
	}, []*types.HistoryEvent{{}, {}, {}})
}

func (s *workflowHandlerSuite) TestGetWorkflowExecutionHistory__Success__RawHistoryEnabledNoTransientDecisionEmitted() {
	var nextEventID int64 = 5
	s.getWorkflowExecutionHistory(5, &types.TransientDecisionInfo{
		StartedEvent:   &types.HistoryEvent{EventID: nextEventID + 1},
		ScheduledEvent: &types.HistoryEvent{EventID: nextEventID},
	}, []*types.HistoryEvent{{}, {}, {}})
}

func (s *workflowHandlerSuite) getWorkflowExecutionHistory(nextEventID int64, transientDecision *types.TransientDecisionInfo, historyEvents []*types.HistoryEvent) {
	wh := s.getWorkflowHandler(
		NewConfig(
			dc.NewCollection(
				dc.NewNopClient(),
				s.mockResource.GetLogger()),
			numHistoryShards,
			false,
			true,
		),
	)
	ctx := context.Background()
	s.mockDomainCache.EXPECT().GetDomainID(gomock.Any()).Return(s.testDomainID, nil).AnyTimes()
	s.mockVersionChecker.EXPECT().SupportsRawHistoryQuery(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	blob, _ := wh.GetPayloadSerializer().SerializeBatchEvents(historyEvents, common.EncodingTypeThriftRW)
	s.mockHistoryV2Mgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(&persistence.ReadRawHistoryBranchResponse{
		HistoryEventBlobs: []*persistence.DataBlob{blob},
		NextPageToken:     []byte{},
	}, nil).Once()
	token, _ := json.Marshal(&getHistoryContinuationToken{
		FirstEventID:      1,
		NextEventID:       nextEventID,
		RunID:             testRunID,
		TransientDecision: transientDecision,
	})
	resp, err := wh.GetWorkflowExecutionHistory(ctx, &types.GetWorkflowExecutionHistoryRequest{
		Domain: s.testDomain,
		Execution: &types.WorkflowExecution{
			WorkflowID: testWorkflowID,
			RunID:      testRunID,
		},
		SkipArchival:  true,
		NextPageToken: token,
	})
	s.NoError(err)
	s.NotNil(resp)
	s.NotNil(resp.RawHistory)
	s.Equal(2, len(resp.RawHistory))

	events := deserializeBlobDataToHistoryEvents(wh, resp.RawHistory)
	s.NotNil(events)
	if transientDecision != nil {
		s.Equal(len(historyEvents)+2, len(events))
	} else {
		s.Equal(len(historyEvents), len(events))
	}
}

func deserializeBlobDataToHistoryEvents(wh *WorkflowHandler, dataBlobs []*types.DataBlob) []*types.HistoryEvent {
	var historyEvents []*types.HistoryEvent
	for _, batch := range dataBlobs {
		events, err := wh.GetPayloadSerializer().DeserializeBatchEvents(&persistence.DataBlob{Data: batch.Data, Encoding: common.EncodingTypeThriftRW})
		if err != nil {
			return nil
		}
		historyEvents = append(historyEvents, events...)
	}
	return historyEvents
}

func (s *workflowHandlerSuite) TestListWorkflowExecutions() {
	config := s.newConfig(dc.NewInMemoryClient())
	wh := s.getWorkflowHandler(config)

	s.mockDomainCache.EXPECT().GetDomainID(gomock.Any()).Return(s.testDomainID, nil).AnyTimes()
	s.mockVisibilityMgr.On("ListWorkflowExecutions", mock.Anything, mock.Anything).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Once()

	listRequest := &types.ListWorkflowExecutionsRequest{
		Domain:   s.testDomain,
		PageSize: int32(config.ESIndexMaxResultWindow()),
	}
	ctx := context.Background()

	query := "WorkflowID = 'wid'"
	listRequest.Query = query
	_, err := wh.ListWorkflowExecutions(ctx, listRequest)
	s.NoError(err)
	s.Equal(query, listRequest.GetQuery())

	query = "InvalidKey = 'a'"
	listRequest.Query = query
	_, err = wh.ListWorkflowExecutions(ctx, listRequest)
	s.NotNil(err)

	listRequest.PageSize = int32(config.ESIndexMaxResultWindow() + 1)
	_, err = wh.ListWorkflowExecutions(ctx, listRequest)
	s.NotNil(err)
}

func (s *workflowHandlerSuite) TestScantWorkflowExecutions() {
	config := s.newConfig(dc.NewInMemoryClient())
	wh := s.getWorkflowHandler(config)

	s.mockDomainCache.EXPECT().GetDomainID(gomock.Any()).Return(s.testDomainID, nil).AnyTimes()
	s.mockVisibilityMgr.On("ScanWorkflowExecutions", mock.Anything, mock.Anything).Return(&persistence.ListWorkflowExecutionsResponse{}, nil).Once()

	listRequest := &types.ListWorkflowExecutionsRequest{
		Domain:   s.testDomain,
		PageSize: int32(config.ESIndexMaxResultWindow()),
	}
	ctx := context.Background()

	query := "WorkflowID = 'wid'"
	listRequest.Query = query
	_, err := wh.ScanWorkflowExecutions(ctx, listRequest)
	s.NoError(err)
	s.Equal(query, listRequest.GetQuery())

	query = "InvalidKey = 'a'"
	listRequest.Query = query
	_, err = wh.ScanWorkflowExecutions(ctx, listRequest)
	s.NotNil(err)

	listRequest.PageSize = int32(config.ESIndexMaxResultWindow() + 1)
	_, err = wh.ListWorkflowExecutions(ctx, listRequest)
	s.NotNil(err)
}

func (s *workflowHandlerSuite) TestCountWorkflowExecutions() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	s.mockDomainCache.EXPECT().GetDomainID(gomock.Any()).Return(s.testDomainID, nil).AnyTimes()
	s.mockVisibilityMgr.On("CountWorkflowExecutions", mock.Anything, mock.Anything).Return(&persistence.CountWorkflowExecutionsResponse{}, nil).Once()

	countRequest := &types.CountWorkflowExecutionsRequest{
		Domain: s.testDomain,
	}
	ctx := context.Background()

	query := "WorkflowID = 'wid'"
	countRequest.Query = query
	_, err := wh.CountWorkflowExecutions(ctx, countRequest)
	s.NoError(err)
	s.Equal(query, countRequest.GetQuery())

	query = "InvalidKey = 'a'"
	countRequest.Query = query
	_, err = wh.CountWorkflowExecutions(ctx, countRequest)
	s.NotNil(err)
}

func (s *workflowHandlerSuite) TestConvertIndexedKeyToThrift() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))
	m := map[string]interface{}{
		"key1":  float64(0),
		"key2":  float64(1),
		"key3":  float64(2),
		"key4":  float64(3),
		"key5":  float64(4),
		"key6":  float64(5),
		"key1i": 0,
		"key2i": 1,
		"key3i": 2,
		"key4i": 3,
		"key5i": 4,
		"key6i": 5,
		"key1t": shared.IndexedValueTypeString,
		"key2t": shared.IndexedValueTypeKeyword,
		"key3t": shared.IndexedValueTypeInt,
		"key4t": shared.IndexedValueTypeDouble,
		"key5t": shared.IndexedValueTypeBool,
		"key6t": shared.IndexedValueTypeDatetime,
	}
	result := wh.convertIndexedKeyToThrift(m)
	s.Equal(types.IndexedValueTypeString, result["key1"])
	s.Equal(types.IndexedValueTypeKeyword, result["key2"])
	s.Equal(types.IndexedValueTypeInt, result["key3"])
	s.Equal(types.IndexedValueTypeDouble, result["key4"])
	s.Equal(types.IndexedValueTypeBool, result["key5"])
	s.Equal(types.IndexedValueTypeDatetime, result["key6"])
	s.Equal(types.IndexedValueTypeString, result["key1i"])
	s.Equal(types.IndexedValueTypeKeyword, result["key2i"])
	s.Equal(types.IndexedValueTypeInt, result["key3i"])
	s.Equal(types.IndexedValueTypeDouble, result["key4i"])
	s.Equal(types.IndexedValueTypeBool, result["key5i"])
	s.Equal(types.IndexedValueTypeDatetime, result["key6i"])
	s.Equal(types.IndexedValueTypeString, result["key1t"])
	s.Equal(types.IndexedValueTypeKeyword, result["key2t"])
	s.Equal(types.IndexedValueTypeInt, result["key3t"])
	s.Equal(types.IndexedValueTypeDouble, result["key4t"])
	s.Equal(types.IndexedValueTypeBool, result["key5t"])
	s.Equal(types.IndexedValueTypeDatetime, result["key6t"])
	s.Panics(func() {
		wh.convertIndexedKeyToThrift(map[string]interface{}{
			"invalidType": "unknown",
		})
	})
}

func (s *workflowHandlerSuite) TestVerifyHistoryIsComplete() {
	events := make([]*types.HistoryEvent, 50)
	for i := 0; i < len(events); i++ {
		events[i] = &types.HistoryEvent{EventID: int64(i + 1)}
	}
	var eventsWithHoles []*types.HistoryEvent
	eventsWithHoles = append(eventsWithHoles, events[9:12]...)
	eventsWithHoles = append(eventsWithHoles, events[20:31]...)

	testCases := []struct {
		events       []*types.HistoryEvent
		firstEventID int64
		lastEventID  int64
		isFirstPage  bool
		isLastPage   bool
		pageSize     int
		isResultErr  bool
	}{
		{events[:1], 1, 1, true, true, 1000, false},
		{events[:5], 1, 5, true, true, 1000, false},
		{events[9:31], 10, 31, true, true, 1000, false},
		{events[9:29], 10, 50, true, false, 20, false},
		{events[9:30], 10, 50, true, false, 20, false},

		{events[9:29], 1, 50, false, false, 20, false},
		{events[9:29], 1, 29, false, true, 20, false},

		{eventsWithHoles, 1, 50, false, false, 22, true},
		{eventsWithHoles, 10, 50, true, false, 22, true},
		{eventsWithHoles, 1, 31, false, true, 22, true},
		{eventsWithHoles, 10, 31, true, true, 1000, true},

		{events[9:31], 9, 31, true, true, 1000, true},
		{events[9:31], 9, 50, true, false, 22, true},
		{events[9:31], 11, 31, true, true, 1000, true},
		{events[9:31], 11, 50, true, false, 22, true},

		{events[9:31], 10, 30, true, true, 1000, true},
		{events[9:31], 1, 30, false, true, 22, true},
		{events[9:31], 10, 32, true, true, 1000, true},
		{events[9:31], 1, 32, false, true, 22, true},
	}

	for i, tc := range testCases {
		err := verifyHistoryIsComplete(tc.events, tc.firstEventID, tc.lastEventID, tc.isFirstPage, tc.isLastPage, tc.pageSize)
		if tc.isResultErr {
			s.Error(err, "testcase %v failed", i)
		} else {
			s.NoError(err, "testcase %v failed", i)
		}
	}
}

func (s *workflowHandlerSuite) TestContextMetricsTags() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	tag := metrics.ThriftTransportTag()
	ctx := metrics.TagContext(context.Background(), tag)
	wh.CountWorkflowExecutions(ctx, nil)

	snapshot := s.mockResource.MetricsScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if counter.Name() == "test.cadence_requests" {
			s.Equal(tag.Value(), counter.Tags()[tag.Key()])
			return
		}
	}
	s.Fail("counter not found")
}

func (s *workflowHandlerSuite) TestSignalMetricHasSignalName() {
	wh := s.getWorkflowHandler(s.newConfig(dc.NewInMemoryClient()))

	signalRequest := &types.SignalWorkflowExecutionRequest{
		SignalName: "test_signal",
	}
	wh.SignalWorkflowExecution(context.Background(), signalRequest)

	expectedMetrics := make(map[string]bool)
	expectedMetrics["test.cadence_requests"] = false
	expectedMetrics["test.cadence_errors_bad_request"] = false

	snapshot := s.mockResource.MetricsScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if _, ok := expectedMetrics[counter.Name()]; ok {
			expectedMetrics[counter.Name()] = true
		}
		if val, ok := counter.Tags()["signalName"]; ok {
			s.Equal(val, "test_signal")
		} else {
			s.Fail("Couldn't find signalName tag")
		}
	}
	s.True(expectedMetrics["test.cadence_requests"])
	s.True(expectedMetrics["test.cadence_errors_bad_request"])
}

func (s *workflowHandlerSuite) newConfig(dynamicClient dc.Client) *Config {
	return NewConfig(
		dc.NewCollection(
			dynamicClient,
			s.mockResource.GetLogger(),
		),
		numHistoryShards,
		false,
		false,
	)
}

func updateRequest(
	historyArchivalURI *string,
	historyArchivalStatus *types.ArchivalStatus,
	visibilityArchivalURI *string,
	visibilityArchivalStatus *types.ArchivalStatus,
) *types.UpdateDomainRequest {
	return &types.UpdateDomainRequest{
		Name:                     "test-name",
		HistoryArchivalStatus:    historyArchivalStatus,
		HistoryArchivalURI:       historyArchivalURI,
		VisibilityArchivalStatus: visibilityArchivalStatus,
		VisibilityArchivalURI:    visibilityArchivalURI,
	}
}

func persistenceGetDomainResponse(historyArchivalState, visibilityArchivalState *domain.ArchivalState) *persistence.GetDomainResponse {
	return &persistence.GetDomainResponse{
		Info: &persistence.DomainInfo{
			ID:          "test-id",
			Name:        "test-name",
			Status:      0,
			Description: "test-description",
			OwnerEmail:  "test-owner-email",
			Data:        make(map[string]string),
		},
		Config: &persistence.DomainConfig{
			Retention:                1,
			EmitMetric:               true,
			HistoryArchivalStatus:    historyArchivalState.Status,
			HistoryArchivalURI:       historyArchivalState.URI,
			VisibilityArchivalStatus: visibilityArchivalState.Status,
			VisibilityArchivalURI:    visibilityArchivalState.URI,
		},
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{
					ClusterName: cluster.TestCurrentClusterName,
				},
			},
		},
		IsGlobalDomain:              false,
		ConfigVersion:               0,
		FailoverVersion:             0,
		FailoverNotificationVersion: 0,
		NotificationVersion:         0,
	}
}

func registerDomainRequest(
	historyArchivalStatus *types.ArchivalStatus,
	historyArchivalURI string,
	visibilityArchivalStatus *types.ArchivalStatus,
	visibilityArchivalURI string,
) *types.RegisterDomainRequest {
	return &types.RegisterDomainRequest{
		Name:                                   "test-domain",
		Description:                            "test-description",
		OwnerEmail:                             "test-owner-email",
		WorkflowExecutionRetentionPeriodInDays: 10,
		EmitMetric:                             common.BoolPtr(true),
		Clusters: []*types.ClusterReplicationConfiguration{
			{
				ClusterName: cluster.TestCurrentClusterName,
			},
		},
		ActiveClusterName:        cluster.TestCurrentClusterName,
		Data:                     make(map[string]string),
		SecurityToken:            "token",
		HistoryArchivalStatus:    historyArchivalStatus,
		HistoryArchivalURI:       historyArchivalURI,
		VisibilityArchivalStatus: visibilityArchivalStatus,
		VisibilityArchivalURI:    visibilityArchivalURI,
		IsGlobalDomain:           false,
	}
}

func getHistoryRequest(nextPageToken []byte) *types.GetWorkflowExecutionHistoryRequest {
	return &types.GetWorkflowExecutionHistoryRequest{
		Execution: &types.WorkflowExecution{
			WorkflowID: testWorkflowID,
			RunID:      testRunID,
		},
		NextPageToken: nextPageToken,
	}
}

func listArchivedWorkflowExecutionsTestRequest() *types.ListArchivedWorkflowExecutionsRequest {
	return &types.ListArchivedWorkflowExecutionsRequest{
		Domain:   "some random domain name",
		PageSize: 10,
		Query:    "some random query string",
	}
}

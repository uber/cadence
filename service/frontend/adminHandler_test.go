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
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
	esmock "github.com/uber/cadence/common/elasticsearch/mocks"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/types"
)

type (
	adminHandlerSuite struct {
		suite.Suite
		*require.Assertions

		controller        *gomock.Controller
		mockResource      *resource.Test
		mockHistoryClient *history.MockClient
		mockDomainCache   *cache.MockDomainCache

		mockHistoryV2Mgr *mocks.HistoryV2Manager

		domainName string
		domainID   string

		handler *adminHandlerImpl
	}
)

func TestAdminHandlerSuite(t *testing.T) {
	s := new(adminHandlerSuite)
	suite.Run(t, s)
}

func (s *adminHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.domainName = "some random domain name"
	s.domainID = "some random domain ID"

	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.controller, metrics.Frontend)
	s.mockDomainCache = s.mockResource.DomainCache
	s.mockHistoryClient = s.mockResource.HistoryClient
	s.mockHistoryV2Mgr = s.mockResource.HistoryMgr

	params := &service.BootstrapParams{
		PersistenceConfig: config.Persistence{
			NumHistoryShards: 1,
		},
	}
	config := &Config{
		EnableAdminProtection:  dynamicconfig.GetBoolPropertyFn(false),
		EnableGracefulFailover: dynamicconfig.GetBoolPropertyFn(false),
	}
	s.handler = NewAdminHandler(s.mockResource, params, config)
	s.handler.Start()
}

func (s *adminHandlerSuite) TearDownTest() {
	s.controller.Finish()
	s.mockResource.Finish(s.T())
	s.handler.Stop()
}

func (s *adminHandlerSuite) Test_ConvertIndexedValueTypeToESDataType() {
	tests := []struct {
		input    types.IndexedValueType
		expected string
	}{
		{
			input:    types.IndexedValueTypeString,
			expected: "text",
		},
		{
			input:    types.IndexedValueTypeKeyword,
			expected: "keyword",
		},
		{
			input:    types.IndexedValueTypeInt,
			expected: "long",
		},
		{
			input:    types.IndexedValueTypeDouble,
			expected: "double",
		},
		{
			input:    types.IndexedValueTypeBool,
			expected: "boolean",
		},
		{
			input:    types.IndexedValueTypeDatetime,
			expected: "date",
		},
		{
			input:    types.IndexedValueType(-1),
			expected: "",
		},
	}

	for _, test := range tests {
		s.Equal(test.expected, convertIndexedValueTypeToESDataType(test.input))
	}
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2_FailedOnInvalidWorkflowID() {

	ctx := context.Background()
	_, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "",
				RunID:      uuid.New(),
			},
			StartEventID:      common.Int64Ptr(1),
			StartEventVersion: common.Int64Ptr(100),
			EndEventID:        common.Int64Ptr(10),
			EndEventVersion:   common.Int64Ptr(100),
			MaximumPageSize:   1,
			NextPageToken:     nil,
		})
	s.Error(err)
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2_FailedOnInvalidRunID() {
	ctx := context.Background()
	_, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "workflowID",
				RunID:      "runID",
			},
			StartEventID:      common.Int64Ptr(1),
			StartEventVersion: common.Int64Ptr(100),
			EndEventID:        common.Int64Ptr(10),
			EndEventVersion:   common.Int64Ptr(100),
			MaximumPageSize:   1,
			NextPageToken:     nil,
		})
	s.Error(err)
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2_FailedOnInvalidSize() {
	ctx := context.Background()
	_, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "workflowID",
				RunID:      uuid.New(),
			},
			StartEventID:      common.Int64Ptr(1),
			StartEventVersion: common.Int64Ptr(100),
			EndEventID:        common.Int64Ptr(10),
			EndEventVersion:   common.Int64Ptr(100),
			MaximumPageSize:   -1,
			NextPageToken:     nil,
		})
	s.Error(err)
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2_FailedOnDomainCache() {
	ctx := context.Background()
	s.mockDomainCache.EXPECT().GetDomainID(s.domainName).Return("", fmt.Errorf("test")).Times(1)
	_, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "workflowID",
				RunID:      uuid.New(),
			},
			StartEventID:      common.Int64Ptr(1),
			StartEventVersion: common.Int64Ptr(100),
			EndEventID:        common.Int64Ptr(10),
			EndEventVersion:   common.Int64Ptr(100),
			MaximumPageSize:   1,
			NextPageToken:     nil,
		})
	s.Error(err)
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2() {
	ctx := context.Background()
	s.mockDomainCache.EXPECT().GetDomainID(s.domainName).Return(s.domainID, nil).AnyTimes()
	branchToken := []byte{1}
	versionHistory := persistence.NewVersionHistory(branchToken, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(int64(10), int64(100)),
	})
	rawVersionHistories := persistence.NewVersionHistories(versionHistory)
	versionHistories := rawVersionHistories.ToInternalType()
	mState := &types.GetMutableStateResponse{
		NextEventID:        11,
		CurrentBranchToken: branchToken,
		VersionHistories:   versionHistories,
	}
	s.mockHistoryClient.EXPECT().GetMutableState(gomock.Any(), gomock.Any()).Return(mState, nil).AnyTimes()

	s.mockHistoryV2Mgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(&persistence.ReadRawHistoryBranchResponse{
		HistoryEventBlobs: []*persistence.DataBlob{},
		NextPageToken:     []byte{},
		Size:              0,
	}, nil)
	_, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "workflowID",
				RunID:      uuid.New(),
			},
			StartEventID:      common.Int64Ptr(1),
			StartEventVersion: common.Int64Ptr(100),
			EndEventID:        common.Int64Ptr(10),
			EndEventVersion:   common.Int64Ptr(100),
			MaximumPageSize:   10,
			NextPageToken:     nil,
		})
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_GetWorkflowExecutionRawHistoryV2_SameStartIDAndEndID() {
	ctx := context.Background()
	s.mockDomainCache.EXPECT().GetDomainID(s.domainName).Return(s.domainID, nil).AnyTimes()
	branchToken := []byte{1}
	versionHistory := persistence.NewVersionHistory(branchToken, []*persistence.VersionHistoryItem{
		persistence.NewVersionHistoryItem(int64(10), int64(100)),
	})
	rawVersionHistories := persistence.NewVersionHistories(versionHistory)
	versionHistories := rawVersionHistories.ToInternalType()
	mState := &types.GetMutableStateResponse{
		NextEventID:        11,
		CurrentBranchToken: branchToken,
		VersionHistories:   versionHistories,
	}
	s.mockHistoryClient.EXPECT().GetMutableState(gomock.Any(), gomock.Any()).Return(mState, nil).AnyTimes()

	resp, err := s.handler.GetWorkflowExecutionRawHistoryV2(ctx,
		&types.GetWorkflowExecutionRawHistoryV2Request{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: "workflowID",
				RunID:      uuid.New(),
			},
			StartEventID:      common.Int64Ptr(10),
			StartEventVersion: common.Int64Ptr(100),
			MaximumPageSize:   1,
			NextPageToken:     nil,
		})
	s.Nil(resp.NextPageToken)
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_SetRequestDefaultValueAndGetTargetVersionHistory_DefinedStartAndEnd() {
	inputStartEventID := int64(1)
	inputStartVersion := int64(10)
	inputEndEventID := int64(100)
	inputEndVersion := int64(11)
	firstItem := persistence.NewVersionHistoryItem(inputStartEventID, inputStartVersion)
	endItem := persistence.NewVersionHistoryItem(inputEndEventID, inputEndVersion)
	versionHistory := persistence.NewVersionHistory([]byte{}, []*persistence.VersionHistoryItem{firstItem, endItem})
	versionHistories := persistence.NewVersionHistories(versionHistory)
	request := &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: "workflowID",
			RunID:      uuid.New(),
		},
		StartEventID:      common.Int64Ptr(inputStartEventID),
		StartEventVersion: common.Int64Ptr(inputStartVersion),
		EndEventID:        common.Int64Ptr(inputEndEventID),
		EndEventVersion:   common.Int64Ptr(inputEndVersion),
		MaximumPageSize:   10,
		NextPageToken:     nil,
	}

	targetVersionHistory, err := s.handler.setRequestDefaultValueAndGetTargetVersionHistory(
		request,
		versionHistories,
	)
	s.Equal(request.GetStartEventID(), inputStartEventID)
	s.Equal(request.GetEndEventID(), inputEndEventID)
	s.Equal(targetVersionHistory, versionHistory)
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_SetRequestDefaultValueAndGetTargetVersionHistory_DefinedEndEvent() {
	inputStartEventID := int64(1)
	inputEndEventID := int64(100)
	inputStartVersion := int64(10)
	inputEndVersion := int64(11)
	firstItem := persistence.NewVersionHistoryItem(inputStartEventID, inputStartVersion)
	targetItem := persistence.NewVersionHistoryItem(inputEndEventID, inputEndVersion)
	versionHistory := persistence.NewVersionHistory([]byte{}, []*persistence.VersionHistoryItem{firstItem, targetItem})
	versionHistories := persistence.NewVersionHistories(versionHistory)
	request := &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: "workflowID",
			RunID:      uuid.New(),
		},
		StartEventID:      nil,
		StartEventVersion: nil,
		EndEventID:        common.Int64Ptr(inputEndEventID),
		EndEventVersion:   common.Int64Ptr(inputEndVersion),
		MaximumPageSize:   10,
		NextPageToken:     nil,
	}

	targetVersionHistory, err := s.handler.setRequestDefaultValueAndGetTargetVersionHistory(
		request,
		versionHistories,
	)
	s.Equal(request.GetStartEventID(), inputStartEventID-1)
	s.Equal(request.GetEndEventID(), inputEndEventID)
	s.Equal(targetVersionHistory, versionHistory)
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_SetRequestDefaultValueAndGetTargetVersionHistory_DefinedStartEvent() {
	inputStartEventID := int64(1)
	inputEndEventID := int64(100)
	inputStartVersion := int64(10)
	inputEndVersion := int64(11)
	firstItem := persistence.NewVersionHistoryItem(inputStartEventID, inputStartVersion)
	targetItem := persistence.NewVersionHistoryItem(inputEndEventID, inputEndVersion)
	versionHistory := persistence.NewVersionHistory([]byte{}, []*persistence.VersionHistoryItem{firstItem, targetItem})
	versionHistories := persistence.NewVersionHistories(versionHistory)
	request := &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: "workflowID",
			RunID:      uuid.New(),
		},
		StartEventID:      common.Int64Ptr(inputStartEventID),
		StartEventVersion: common.Int64Ptr(inputStartVersion),
		EndEventID:        nil,
		EndEventVersion:   nil,
		MaximumPageSize:   10,
		NextPageToken:     nil,
	}

	targetVersionHistory, err := s.handler.setRequestDefaultValueAndGetTargetVersionHistory(
		request,
		versionHistories,
	)
	s.Equal(request.GetStartEventID(), inputStartEventID)
	s.Equal(request.GetEndEventID(), inputEndEventID+1)
	s.Equal(targetVersionHistory, versionHistory)
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_SetRequestDefaultValueAndGetTargetVersionHistory_NonCurrentBranch() {
	inputStartEventID := int64(1)
	inputEndEventID := int64(100)
	inputStartVersion := int64(10)
	inputEndVersion := int64(101)
	item1 := persistence.NewVersionHistoryItem(inputStartEventID, inputStartVersion)
	item2 := persistence.NewVersionHistoryItem(inputEndEventID, inputEndVersion)
	versionHistory1 := persistence.NewVersionHistory([]byte{}, []*persistence.VersionHistoryItem{item1, item2})
	item3 := persistence.NewVersionHistoryItem(int64(10), int64(20))
	item4 := persistence.NewVersionHistoryItem(int64(20), int64(51))
	versionHistory2 := persistence.NewVersionHistory([]byte{}, []*persistence.VersionHistoryItem{item1, item3, item4})
	versionHistories := persistence.NewVersionHistories(versionHistory1)
	_, _, err := versionHistories.AddVersionHistory(versionHistory2)
	s.NoError(err)
	request := &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: "workflowID",
			RunID:      uuid.New(),
		},
		StartEventID:      common.Int64Ptr(9),
		StartEventVersion: common.Int64Ptr(20),
		EndEventID:        common.Int64Ptr(inputEndEventID),
		EndEventVersion:   common.Int64Ptr(inputEndVersion),
		MaximumPageSize:   10,
		NextPageToken:     nil,
	}

	targetVersionHistory, err := s.handler.setRequestDefaultValueAndGetTargetVersionHistory(
		request,
		versionHistories,
	)
	s.Equal(request.GetStartEventID(), inputStartEventID)
	s.Equal(request.GetEndEventID(), inputEndEventID)
	s.Equal(targetVersionHistory, versionHistory1)
	s.NoError(err)
}

func (s *adminHandlerSuite) Test_AddSearchAttribute_Validate() {
	handler := s.handler
	handler.params = &service.BootstrapParams{}
	ctx := context.Background()

	type test struct {
		Name     string
		Request  *types.AddSearchAttributeRequest
		Expected error
	}
	// request validation tests
	testCases1 := []test{
		{
			Name:     "nil request",
			Request:  nil,
			Expected: &types.BadRequestError{Message: "Request is nil."},
		},
		{
			Name:     "empty request",
			Request:  &types.AddSearchAttributeRequest{},
			Expected: &types.BadRequestError{Message: "SearchAttributes are not provided"},
		},
		{
			Name: "no advanced config",
			Request: &types.AddSearchAttributeRequest{
				SearchAttribute: map[string]types.IndexedValueType{
					"CustomKeywordField": 1,
				},
			},
			Expected: &types.BadRequestError{Message: "AdvancedVisibilityStore is not configured for this Cadence Cluster"},
		},
	}
	for _, testCase := range testCases1 {
		s.Equal(testCase.Expected, handler.AddSearchAttribute(ctx, testCase.Request))
	}

	dynamicConfig := dynamicconfig.NewMockClient(s.controller)
	handler.params.DynamicConfig = dynamicConfig
	// add advanced visibility store related config
	handler.params.ESConfig = &config.ElasticSearchConfig{}
	esClient := &esmock.GenericClient{}
	defer func() { esClient.AssertExpectations(s.T()) }()
	handler.params.ESClient = esClient
	handler.esClient = esClient

	mockValidAttr := map[string]interface{}{
		"testkey": types.IndexedValueTypeKeyword,
	}
	dynamicConfig.EXPECT().GetMapValue(dynamicconfig.ValidSearchAttributes, nil, definition.GetDefaultIndexedKeys()).
		Return(mockValidAttr, nil).AnyTimes()

	testCases2 := []test{
		{
			Name: "reserved key",
			Request: &types.AddSearchAttributeRequest{
				SearchAttribute: map[string]types.IndexedValueType{
					"WorkflowID": 1,
				},
			},
			Expected: &types.BadRequestError{Message: "Key [WorkflowID] is reserved by system"},
		},
		{
			Name: "key already whitelisted",
			Request: &types.AddSearchAttributeRequest{
				SearchAttribute: map[string]types.IndexedValueType{
					"testkey": 1,
				},
			},
			Expected: &types.BadRequestError{Message: "Key [testkey] is already whitelist"},
		},
	}
	for _, testCase := range testCases2 {
		s.Equal(testCase.Expected, handler.AddSearchAttribute(ctx, testCase.Request))
	}

	dcUpdateTest := test{
		Name: "dynamic config update failed",
		Request: &types.AddSearchAttributeRequest{
			SearchAttribute: map[string]types.IndexedValueType{
				"testkey2": 1,
			},
		},
		Expected: &types.InternalServiceError{Message: "Failed to update dynamic config, err: error"},
	}
	dynamicConfig.EXPECT().UpdateValue(dynamicconfig.ValidSearchAttributes, map[string]interface{}{
		"testkey":  types.IndexedValueTypeKeyword,
		"testkey2": 1,
	}).Return(errors.New("error"))
	s.Equal(dcUpdateTest.Expected, handler.AddSearchAttribute(ctx, dcUpdateTest.Request))

	// ES operations tests
	dynamicConfig.EXPECT().UpdateValue(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	convertFailedTest := test{
		Name: "unknown value type",
		Request: &types.AddSearchAttributeRequest{
			SearchAttribute: map[string]types.IndexedValueType{
				"testkey3": -1,
			},
		},
		Expected: &types.BadRequestError{Message: "Unknown value type, IndexedValueType(-1)"},
	}
	s.Equal(convertFailedTest.Expected, handler.AddSearchAttribute(ctx, convertFailedTest.Request))

	esClient.On("PutMapping", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("error"))
	esClient.On("IsNotFoundError", mock.Anything).Return(false)
	esErrorTest := test{
		Name: "es error",
		Request: &types.AddSearchAttributeRequest{
			SearchAttribute: map[string]types.IndexedValueType{
				"testkey4": 1,
			},
		},
		Expected: &types.InternalServiceError{Message: "Failed to update ES mapping, err: error"},
	}
	s.Equal(esErrorTest.Expected, handler.AddSearchAttribute(ctx, esErrorTest.Request))
}

func (s *adminHandlerSuite) Test_AddSearchAttribute_Permission() {
	ctx := context.Background()
	handler := s.handler
	handler.config = &Config{
		EnableAdminProtection: dynamicconfig.GetBoolPropertyFn(true),
		AdminOperationToken:   dynamicconfig.GetStringPropertyFn(common.DefaultAdminOperationToken),
	}

	type test struct {
		Name     string
		Request  *types.AddSearchAttributeRequest
		Expected error
	}
	testCases := []test{
		{
			Name: "unknown token",
			Request: &types.AddSearchAttributeRequest{
				SecurityToken: "unknown",
			},
			Expected: errNoPermission,
		},
		{
			Name: "correct token",
			Request: &types.AddSearchAttributeRequest{
				SecurityToken: common.DefaultAdminOperationToken,
			},
			Expected: &types.BadRequestError{Message: "SearchAttributes are not provided"},
		},
	}
	for _, testCase := range testCases {
		s.Equal(testCase.Expected, handler.AddSearchAttribute(ctx, testCase.Request))
	}
}

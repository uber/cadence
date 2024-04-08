// Copyright (c) 2024 Uber Technologies, Inc.
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

package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

// newTestHandler creates a new instance of the handler with mocked dependencies for testing.
func newTestHandler(domainManager persistence.DomainManager, primaryCluster bool, domainReplicator Replicator) Handler {
	mockDC := dynamicconfig.NewCollection(dynamicconfig.NewNopClient(), log.NewNoop())
	domainDefaults := &config.ArchivalDomainDefaults{
		History: config.HistoryArchivalDomainDefaults{
			Status: "Disabled",
			URI:    "https://history.example.com",
		},
		Visibility: config.VisibilityArchivalDomainDefaults{
			Status: "Disabled",
			URI:    "https://visibility.example.com",
		},
	}
	archivalMetadata := archiver.NewArchivalMetadata(mockDC, "Enabled", true, "Enabled", true, domainDefaults)
	testConfig := Config{
		MinRetentionDays:       dynamicconfig.GetIntPropertyFn(1),
		MaxRetentionDays:       dynamicconfig.GetIntPropertyFn(5),
		RequiredDomainDataKeys: nil,
		MaxBadBinaryCount:      func(string) int { return 3 },
		FailoverCoolDown:       func(string) time.Duration { return time.Second },
	}

	return NewHandler(
		testConfig,
		log.NewNoop(),
		domainManager,
		cluster.GetTestClusterMetadata(primaryCluster),
		domainReplicator,
		archivalMetadata,
		provider.NewArchiverProvider(nil, nil),
		clock.NewMockedTimeSource(),
	)
}

func TestRegisterDomain(t *testing.T) {
	tests := []struct {
		name             string
		request          *types.RegisterDomainRequest
		isPrimaryCluster bool
		mockSetup        func(*persistence.MockDomainManager, *MockReplicator, *types.RegisterDomainRequest)
		wantErr          bool
		expectedErr      error
	}{
		{
			name:             "success",
			isPrimaryCluster: true,
			request: &types.RegisterDomainRequest{
				Name:                                   "test-domain",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
				mockDomainMgr.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(&persistence.CreateDomainResponse{ID: "test-domain-id"}, nil)
			},
			wantErr: false,
		},
		{
			name:             "domain already exists",
			isPrimaryCluster: true,
			request: &types.RegisterDomainRequest{
				Name:                                   "existing-domain",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(&persistence.GetDomainResponse{}, nil)
			},
			wantErr: true,
		},
		{
			name:             "global domain registration on non-primary cluster",
			isPrimaryCluster: false,
			request: &types.RegisterDomainRequest{
				Name:           "global-test-domain",
				IsGlobalDomain: true,
			},
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().
					GetDomain(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, req *persistence.GetDomainRequest) (*persistence.GetDomainResponse, error) {
						if req.Name == "global-test-domain" {
							return nil, &types.EntityNotExistsError{}
						}
						return nil, errors.New("unexpected domain name")
					}).AnyTimes()

			},
			wantErr:     true,
			expectedErr: errNotPrimaryCluster,
		},
		{
			name: "unexpected error on domain lookup",
			request: &types.RegisterDomainRequest{
				Name:                                   "test-domain-with-lookup-error",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				unexpectedErr := &types.InternalServiceError{Message: "Internal server error."}
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, unexpectedErr)
			},
			wantErr:     true,
			expectedErr: &types.InternalServiceError{},
		},
		{
			name: "domain name does not match regex",
			request: &types.RegisterDomainRequest{
				Name:                                   "invalid_domain!",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: errInvalidDomainName,
		},
		{
			name: "specify active cluster name",
			request: &types.RegisterDomainRequest{
				Name:                                   "test-domain-with-active-cluster",
				WorkflowExecutionRetentionPeriodInDays: 3,
				ActiveClusterName:                      "active",
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})

				mockDomainMgr.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *persistence.CreateDomainRequest) (*persistence.CreateDomainResponse, error) {
						return &persistence.CreateDomainResponse{ID: "test-domain-id"}, nil
					})
			},
			wantErr: false,
		},
		{
			name: "specify clusters including an invalid one",
			request: &types.RegisterDomainRequest{
				Name: "test-domain-with-clusters",
				Clusters: []*types.ClusterReplicationConfiguration{
					{ClusterName: "valid-cluster-1"},
					{ClusterName: "invalid-cluster"},
				},
				IsGlobalDomain: true,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &types.BadRequestError{},
		},
		{
			name: "invalid history archival configuration",
			request: &types.RegisterDomainRequest{
				Name:                  "test-domain-invalid-archival-config",
				HistoryArchivalStatus: types.ArchivalStatusEnabled.Ptr(),
				HistoryArchivalURI:    "invalid-uri",
				IsGlobalDomain:        true,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &url.Error{},
		},
		{
			name: "error during domain creation",
			request: &types.RegisterDomainRequest{
				Name:                                   "domain-creation-error",
				WorkflowExecutionRetentionPeriodInDays: 2,
				IsGlobalDomain:                         false,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})

				mockDomainMgr.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(nil, errors.New("creation failed"))
			},
			wantErr: true,
		},
		{
			name: "global domain with replication task",
			request: &types.RegisterDomainRequest{
				Name:                                   "global-domain-with-replication",
				IsGlobalDomain:                         true,
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, replicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
				mockDomainMgr.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(&persistence.CreateDomainResponse{ID: "domain-id"}, nil)
				replicator.EXPECT().HandleTransmissionTask(gomock.Any(), types.DomainOperationCreate, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), common.InitialPreviousFailoverVersion, true).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "global domain replication task failure",
			request: &types.RegisterDomainRequest{
				Name:                                   "global-domain-replication-failure",
				IsGlobalDomain:                         true,
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, mockReplicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
				mockDomainMgr.EXPECT().CreateDomain(gomock.Any(), gomock.Any()).Return(&persistence.CreateDomainResponse{ID: "domain-id"}, nil)
				mockReplicator.EXPECT().HandleTransmissionTask(
					gomock.Any(),
					types.DomainOperationCreate,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					common.InitialPreviousFailoverVersion,
					true,
				).Return(errors.New("replication task failed"))
			},
			wantErr:     true,
			expectedErr: errors.New("replication task failed"),
		},
		{
			name: "visibility archival URI triggers parsing/validation error and not able to reach next state",
			request: &types.RegisterDomainRequest{
				Name:                     "domain-with-invalid-visibility-uri",
				VisibilityArchivalStatus: types.ArchivalStatusEnabled.Ptr(),
				VisibilityArchivalURI:    "invalid-visibility-uri",
				IsGlobalDomain:           true,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, mockReplicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &url.Error{},
		},
		{
			name: "local domain with invalid replication configuration: non-global domain",
			request: &types.RegisterDomainRequest{
				Name:              "local-invalid-replication",
				IsGlobalDomain:    false,
				ActiveClusterName: "current-cluster",
				Clusters: []*types.ClusterReplicationConfiguration{
					{ClusterName: "non-current-cluster"},
					{ClusterName: "non-current-cluster2"},
				},
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, mockReplicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &types.BadRequestError{Message: "Invalid local domain active cluster"},
		},
		{
			name: "local domain with invalid replication configuration: global domain",
			request: &types.RegisterDomainRequest{
				Name:              "global-invalid-replication",
				IsGlobalDomain:    true,
				ActiveClusterName: "current-cluster",
				Clusters: []*types.ClusterReplicationConfiguration{
					{ClusterName: "non-current-cluster2"},
					{ClusterName: "non-current-cluster3"},
				},
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			isPrimaryCluster: true,
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, mockReplicator *MockReplicator, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &types.BadRequestError{Message: "Invalid local domain active cluster"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)

			mockDomainMgr := persistence.NewMockDomainManager(controller)
			mockReplicator := NewMockReplicator(controller)

			handler := newTestHandler(mockDomainMgr, tc.isPrimaryCluster, mockReplicator)

			tc.mockSetup(mockDomainMgr, mockReplicator, tc.request)

			err := handler.RegisterDomain(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)

				if tc.expectedErr != nil {
					assert.IsType(t, tc.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListDomains(t *testing.T) {

	var configIsolationGroup types.IsolationGroupConfiguration
	tests := []struct {
		name          string
		request       *types.ListDomainsRequest
		setupMocks    func(mockDomainManager *persistence.MockDomainManager)
		expectedResp  *types.ListDomainsResponse
		expectedError error
	}{
		{
			name: "Success case - List domains",
			request: &types.ListDomainsRequest{
				PageSize: 2,
			},
			setupMocks: func(mockDomainManager *persistence.MockDomainManager) {
				resp := &persistence.ListDomainsResponse{
					Domains: []*persistence.GetDomainResponse{
						{
							Info: &persistence.DomainInfo{
								ID:   "domainID1",
								Name: "domainName1",
							},
							Config:            &persistence.DomainConfig{},
							ReplicationConfig: &persistence.DomainReplicationConfig{},
						},
						{
							Info: &persistence.DomainInfo{
								ID:   "domainID2",
								Name: "domainName2",
							},
							Config:            &persistence.DomainConfig{},
							ReplicationConfig: &persistence.DomainReplicationConfig{},
						},
					},
					NextPageToken: nil,
				}
				mockDomainManager.EXPECT().ListDomains(gomock.Any(), &persistence.ListDomainsRequest{
					PageSize:      2,
					NextPageToken: nil,
				}).Return(resp, nil)
			},
			expectedResp: &types.ListDomainsResponse{
				Domains: []*types.DescribeDomainResponse{
					{
						DomainInfo: &types.DomainInfo{
							Name:        "domainName1",
							UUID:        "domainID1",
							Status:      types.DomainStatusRegistered.Ptr(),
							Description: "",
							OwnerEmail:  "",
							Data:        nil,
						},
						Configuration: &types.DomainConfiguration{
							EmitMetric:               false,
							BadBinaries:              &types.BadBinaries{Binaries: nil},
							HistoryArchivalStatus:    types.ArchivalStatusDisabled.Ptr(),
							HistoryArchivalURI:       "",
							VisibilityArchivalStatus: types.ArchivalStatusDisabled.Ptr(),
							VisibilityArchivalURI:    "",
							IsolationGroups:          &configIsolationGroup,
							AsyncWorkflowConfig:      &types.AsyncWorkflowConfiguration{},
						},
						ReplicationConfiguration: &types.DomainReplicationConfiguration{
							ActiveClusterName: "",
							Clusters:          []*types.ClusterReplicationConfiguration{},
						},
					},
					{
						DomainInfo: &types.DomainInfo{
							Name:        "domainName2",
							UUID:        "domainID2",
							Status:      types.DomainStatusRegistered.Ptr(),
							Description: "",
							OwnerEmail:  "",
							Data:        nil,
						},
						Configuration: &types.DomainConfiguration{
							EmitMetric:               false,
							BadBinaries:              &types.BadBinaries{Binaries: nil},
							HistoryArchivalStatus:    types.ArchivalStatusDisabled.Ptr(),
							HistoryArchivalURI:       "",
							VisibilityArchivalStatus: types.ArchivalStatusDisabled.Ptr(),
							VisibilityArchivalURI:    "",
							IsolationGroups:          &configIsolationGroup,
							AsyncWorkflowConfig:      &types.AsyncWorkflowConfiguration{},
						},
						ReplicationConfiguration: &types.DomainReplicationConfiguration{
							ActiveClusterName: "",
							Clusters:          []*types.ClusterReplicationConfiguration{},
						},
					},
				},
				NextPageToken: nil,
			},
			expectedError: nil,
		},
		{
			name: "Error case - Persistence error",
			request: &types.ListDomainsRequest{
				PageSize: 10,
			},
			setupMocks: func(mockDomainManager *persistence.MockDomainManager) {
				mockDomainManager.EXPECT().ListDomains(gomock.Any(), &persistence.ListDomainsRequest{
					PageSize:      10,
					NextPageToken: nil,
				}).Return(nil, errors.New("persistence error"))
			},
			expectedResp:  nil,
			expectedError: errors.New("persistence error"),
		},
	}

	for _, test := range tests {
		controller := gomock.NewController(t)
		mockDomainMgr := persistence.NewMockDomainManager(controller)
		mockReplicator := NewMockReplicator(controller)
		handler := newTestHandler(mockDomainMgr, true, mockReplicator)

		t.Run(test.name, func(t *testing.T) {
			test.setupMocks(mockDomainMgr)

			resp, err := handler.ListDomains(context.Background(), test.request)

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, test.expectedResp, resp)
			}
		})

	}
}

func TestHandler_DescribeDomain(t *testing.T) {
	controller := gomock.NewController(t)
	mockDomainManager := persistence.NewMockDomainManager(controller)
	mockReplicator := NewMockReplicator(controller)

	handler := newTestHandler(mockDomainManager, true, mockReplicator)

	domainName := "test-domain"
	domainID := "test-domain-id"
	isGlobalDomain := true
	failoverVersion := int64(123)
	lastUpdatedTime := time.Now().UnixNano()
	failoverEndTime := lastUpdatedTime + int64(time.Hour)
	var configIsolationGroup types.IsolationGroupConfiguration

	domainInfo := &persistence.DomainInfo{
		ID:          domainID,
		Name:        domainName,
		Status:      persistence.DomainStatusRegistered,
		Description: "Test Domain",
		OwnerEmail:  "test@uber.com",
		Data:        map[string]string{"k1": "v1"},
	}
	domainConfig := &persistence.DomainConfig{
		Retention: 24,
	}
	domainReplicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: "active",
		Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
	}

	mockDomainManager.EXPECT().
		GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: domainName}).
		Return(&persistence.GetDomainResponse{
			Info:              domainInfo,
			Config:            domainConfig,
			ReplicationConfig: domainReplicationConfig,
			IsGlobalDomain:    isGlobalDomain,
			FailoverVersion:   failoverVersion,
			FailoverEndTime:   &failoverEndTime,
			LastUpdatedTime:   lastUpdatedTime,
		}, nil)

	describeRequest := &types.DescribeDomainRequest{Name: &domainName}
	expectedResponse := &types.DescribeDomainResponse{
		IsGlobalDomain:  isGlobalDomain,
		FailoverVersion: failoverVersion,
		FailoverInfo: &types.FailoverInfo{
			FailoverVersion:         failoverVersion,
			FailoverStartTimestamp:  lastUpdatedTime,
			FailoverExpireTimestamp: failoverEndTime,
		},
		DomainInfo: &types.DomainInfo{
			Name:        domainName,
			Status:      types.DomainStatusRegistered.Ptr(),
			Description: "Test Domain",
			OwnerEmail:  "test@uber.com",
			Data:        map[string]string{"k1": "v1"},
			UUID:        domainID,
		},
		Configuration: &types.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: 24,
			EmitMetric:                             false,
			BadBinaries:                            &types.BadBinaries{Binaries: nil},
			HistoryArchivalStatus:                  types.ArchivalStatusDisabled.Ptr(),
			HistoryArchivalURI:                     "",
			VisibilityArchivalStatus:               types.ArchivalStatusDisabled.Ptr(),
			VisibilityArchivalURI:                  "",
			IsolationGroups:                        &configIsolationGroup,
			AsyncWorkflowConfig:                    &types.AsyncWorkflowConfiguration{},
		},
		ReplicationConfiguration: &types.DomainReplicationConfiguration{
			ActiveClusterName: "active",
			Clusters: []*types.ClusterReplicationConfiguration{
				{ClusterName: "active"},
				{ClusterName: "standby"},
			},
		},
	}

	resp, err := handler.DescribeDomain(context.Background(), describeRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedResponse, resp)

	mockDomainManager.EXPECT().
		GetDomain(gomock.Any(), gomock.Any()).
		Return(nil, types.EntityNotExistsError{}).
		Times(1)

	_, err = handler.DescribeDomain(context.Background(), describeRequest)

	assert.Error(t, types.EntityNotExistsError{})
}

func TestHandler_DeprecateDomain(t *testing.T) {
	tests := []struct {
		name           string
		domainName     string
		setupMocks     func(m *persistence.MockDomainManager, r *MockReplicator)
		isGlobalDomain bool
		primaryCluster bool
		expectedErr    error
	}{
		{
			name:       "success - deprecate local domain",
			domainName: "local-domain",
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				// Mock calls
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: "local-domain"}).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "local-domain", Status: persistence.DomainStatusRegistered},
					Config: &persistence.DomainConfig{
						Retention: 7,
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    false,
				}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
			},
			isGlobalDomain: false,
			primaryCluster: true,
			expectedErr:    nil,
		},
		{
			name:       "failure - domain not found",
			domainName: "non-existent-domain",
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: "non-existent-domain"}).Return(nil, errors.New("domain not found"))
			},
			isGlobalDomain: false,
			primaryCluster: false,
			expectedErr:    errors.New("domain not found"),
		},
		{
			name:       "failure - get metadata error",
			domainName: "test-domain",
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(nil, errors.New("metadata error"))
			},
			expectedErr: errors.New("metadata error"),
		},
		{
			name:           "failure - not primary cluster for global domain",
			domainName:     "global-domain",
			isGlobalDomain: true,
			primaryCluster: false,
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: "global-domain"}).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-domain", Status: persistence.DomainStatusRegistered},
					Config: &persistence.DomainConfig{
						Retention: 7,
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    true,
				}, nil)
			},
			expectedErr: errNotPrimaryCluster,
		},
		{
			name:           "failure - update domain error",
			domainName:     "test-domain",
			primaryCluster: true,
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "test-domain", Status: persistence.DomainStatusRegistered},
					Config: &persistence.DomainConfig{
						Retention: 7,
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					FailoverVersion:   123,
					IsGlobalDomain:    true,
					ConfigVersion:     1,
				}, nil)

				m.EXPECT().UpdateDomain(gomock.Any(), gomock.AssignableToTypeOf(&persistence.UpdateDomainRequest{})).DoAndReturn(
					func(_ context.Context, req *persistence.UpdateDomainRequest) error {
						if req.Info.Name != "test-domain" || req.ConfigVersion != 2 || req.Info.Status != persistence.DomainStatusDeprecated {
							return errors.New("unexpected UpdateDomainRequest")
						}
						return errors.New("update domain error")
					},
				)
			},
			expectedErr: errors.New("update domain error"),
		},
		{
			name:           "failure - replicator error for global domain",
			domainName:     "global-domain",
			isGlobalDomain: true,
			primaryCluster: true,
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info:           &persistence.DomainInfo{Status: persistence.DomainStatusDeprecated},
					IsGlobalDomain: true,
					ConfigVersion:  1}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
				r.EXPECT().HandleTransmissionTask(gomock.Any(), types.DomainOperationUpdate, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true).Return(errors.New("replicator error"))
			},
			expectedErr: errors.New("replicator error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)

			mockDomainManager := persistence.NewMockDomainManager(controller)
			mockReplicator := NewMockReplicator(controller)

			handler := newTestHandler(mockDomainManager, tc.primaryCluster, mockReplicator)
			tc.setupMocks(mockDomainManager, mockReplicator)

			err := handler.DeprecateDomain(context.Background(), &types.DeprecateDomainRequest{Name: tc.domainName})

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_UpdateIsolationGroups(t *testing.T) {
	tests := []struct {
		name             string
		domain           string
		isolationGroups  types.IsolationGroupConfiguration
		isPrimaryCluster bool
		setupMocks       func(m *persistence.MockDomainManager, r *MockReplicator)
		expectedErr      error
	}{
		{
			name:   "Successful Update for Local Domain",
			domain: "local-domain",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {Name: "group1", State: types.IsolationGroupStateHealthy},
			},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "local-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{
						Name: "local-domain",
						ID:   "domainID",
					},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						IsolationGroups:          types.IsolationGroupConfiguration{"group1": {Name: "group1", State: types.IsolationGroupStateHealthy}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  false,
					LastUpdatedTime: time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC).UnixNano(),
				}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
			},

			expectedErr: nil,
		},
		{
			name:   "Successful Update for Global Domain",
			domain: "global-domain",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {State: types.IsolationGroupStateDrained},
			},
			isPrimaryCluster: true,
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-domain"},
					Config: &persistence.DomainConfig{
						Retention: 30,
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					ConfigVersion:     int64(1),
					IsGlobalDomain:    true,
					LastUpdatedTime:   time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC).UnixNano(),
				}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
				r.EXPECT().HandleTransmissionTask(gomock.Any(), types.DomainOperationUpdate, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Failure Due to Invalid Request (nil Isolation Groups)",
			domain: "domain-with-nil-isolation-groups",
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
			},
			expectedErr: fmt.Errorf("invalid request, isolationGroup configuration must be set: Got: {domain-with-nil-isolation-groups map[]}"),
		},
		{
			name:   "Failure Due to Domain Not Found",
			domain: "nonexistent-domain",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {State: types.IsolationGroupStateHealthy},
			},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(nil, errors.New("domain not found"))
			},
			expectedErr: errors.New("domain not found"),
		},
		{
			name:   "Failure Due to UpdateDomain Error",
			domain: "domain-with-update-error",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {State: types.IsolationGroupStateHealthy},
			},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info:              &persistence.DomainInfo{Name: "domain-with-update-error"},
					Config:            &persistence.DomainConfig{},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    false,
				}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(errors.New("update error"))
			},
			expectedErr: errors.New("update error"),
		},
		{
			name:   "Failure Due to Replication Error (for Global Domains)",
			domain: "global-domain-with-replication-error",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {State: types.IsolationGroupStateHealthy},
			},
			isPrimaryCluster: true,
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info:              &persistence.DomainInfo{Name: "global-domain-with-replication-error"},
					Config:            &persistence.DomainConfig{},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    true,
				}, nil)
				m.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
				r.EXPECT().HandleTransmissionTask(gomock.Any(), types.DomainOperationUpdate, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true).Return(errors.New("replication error"))
			},
			expectedErr: errors.New("replication error"),
		},
		{
			name:   "Failure due to GetMetadata error",
			domain: "test-domain",
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(nil, errors.New("metadata error"))
			},
			expectedErr: errors.New("metadata error"),
		},
		{
			name:            "Failure due to nil domain config",
			domain:          "test-domain",
			isolationGroups: types.IsolationGroupConfiguration{},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(&persistence.GetDomainResponse{
					Info:              &persistence.DomainInfo{Name: "test-domain", ID: "domainID"},
					Config:            nil,
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    false,
				}, nil)
			},
			expectedErr: fmt.Errorf("unable to load config for domain as expected"),
		},
		{
			name:   "Failure due to failover cool-down",
			domain: "test-domain",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {Name: "group1", State: types.IsolationGroupStateHealthy},
			},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				lastUpdatedWithinCoolDownPeriod := time.Now().UnixNano() - (500 * int64(time.Millisecond))
				m.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  false,
					LastUpdatedTime: lastUpdatedWithinCoolDownPeriod,
				}, nil)
			},
			expectedErr: errDomainUpdateTooFrequent,
		},
		{
			name:   "Global domain update from non-primary cluster",
			domain: "global-domain",
			isolationGroups: types.IsolationGroupConfiguration{
				"group1": {Name: "group1", State: types.IsolationGroupStateHealthy},
			},
			setupMocks: func(m *persistence.MockDomainManager, r *MockReplicator) {
				m.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				m.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "global-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention: 7,
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{},
					IsGlobalDomain:    true,
				}, nil)
			},
			expectedErr: errNotPrimaryCluster,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)

			mockDomainManager := persistence.NewMockDomainManager(controller)
			mockReplicator := NewMockReplicator(controller)

			handler := newTestHandler(mockDomainManager, tc.isPrimaryCluster, mockReplicator)
			tc.setupMocks(mockDomainManager, mockReplicator)

			err := handler.UpdateIsolationGroups(context.Background(), types.UpdateDomainIsolationGroupsRequest{
				Domain:          tc.domain,
				IsolationGroups: tc.isolationGroups,
			})

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_UpdateAsyncWorkflowConfiguraton(t *testing.T) {

	testDomain := "testDomain"

	tests := []struct {
		name             string
		setupMocks       func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator)
		isPrimaryCluster bool
		request          *types.UpdateDomainAsyncWorkflowConfiguratonRequest
		expectedErr      error
	}{
		{
			name: "fail to get metadata",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(nil, fmt.Errorf("metadata error"))
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: testDomain,
			},
			expectedErr: fmt.Errorf("metadata error"),
		},
		{
			name: "fail to get domain",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("get domain error"))
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: testDomain,
			},
			expectedErr: fmt.Errorf("get domain error"),
		},
		{
			name: "success",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: testDomain}).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{ID: "domainID", Name: testDomain},
					Config: &persistence.DomainConfig{
						Retention:                10,
						EmitMetric:               true,
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://history.example.com",
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						VisibilityArchivalURI:    "https://visibility.example.com",
						BadBinaries:              types.BadBinaries{},
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{},
					},
					ConfigVersion:   1,
					FailoverVersion: 1,
				}, nil)
				mockDomainManager.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: testDomain,
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: nil,
		},
		{
			name: "fail due to nil domain config",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: testDomain}).Return(&persistence.GetDomainResponse{
					Info:   &persistence.DomainInfo{ID: "domainID", Name: testDomain},
					Config: nil, // Domain config is nil
				}, nil)
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: testDomain,
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: fmt.Errorf("unable to load config for domain as expected"),
		},
		{
			name: "fail due to update too frequent",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				lastUpdatedWithinCoolDownPeriod := time.Now().UnixNano() - (500 * int64(time.Millisecond))
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  false,
					LastUpdatedTime: lastUpdatedWithinCoolDownPeriod,
				}, nil)
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "test-domain",
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: errDomainUpdateTooFrequent,
		},
		{
			name:             "fail due to non-primary cluster update for global domain",
			isPrimaryCluster: false,
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "global-test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  true,
					LastUpdatedTime: time.Now().Add(-24 * time.Hour).UnixNano(),
				}, nil)

			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "global-test-domain",
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: errNotPrimaryCluster,
		},
		{
			name: "configuration delete request",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)

				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: true},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  false,
					LastUpdatedTime: time.Now().Add(-24 * time.Hour).UnixNano(),
				}, nil)

				mockDomainManager.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, request *persistence.UpdateDomainRequest) error {
						require.Empty(t, request.Config.AsyncWorkflowConfig)
						return nil
					},
				)
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain:        "test-domain",
				Configuration: nil,
			},
			expectedErr: nil,
		},
		{
			name: "fail on domain manager UpdateDomain call",
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)

				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  false,
					LastUpdatedTime: time.Now().Add(-24 * time.Hour).UnixNano(),
				}, nil)

				mockDomainManager.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update domain error"))
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "test-domain",
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: fmt.Errorf("update domain error"),
		},
		{
			name:             "global domain update triggers replication",
			isPrimaryCluster: true,
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "global-test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  true,
					LastUpdatedTime: time.Now().Add(-24 * time.Hour).UnixNano(),
				}, nil)

				mockDomainManager.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)

				mockReplicator.EXPECT().HandleTransmissionTask(
					gomock.Any(),
					types.DomainOperationUpdate,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					true,
				).Return(nil)
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "global-test-domain",
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: nil,
		},
		{
			name:             "global domain update replication error",
			isPrimaryCluster: true,
			setupMocks: func(mockDomainManager *persistence.MockDomainManager, mockReplicator *MockReplicator) {
				mockDomainManager.EXPECT().GetMetadata(gomock.Any()).Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
				mockDomainManager.EXPECT().GetDomain(gomock.Any(), gomock.Eq(&persistence.GetDomainRequest{Name: "global-test-domain"})).Return(&persistence.GetDomainResponse{
					Info: &persistence.DomainInfo{Name: "global-test-domain", ID: "domainID"},
					Config: &persistence.DomainConfig{
						Retention:                7,
						EmitMetric:               true,
						BadBinaries:              types.BadBinaries{Binaries: map[string]*types.BadBinaryInfo{}},
						HistoryArchivalStatus:    types.ArchivalStatusEnabled,
						VisibilityArchivalStatus: types.ArchivalStatusEnabled,
						HistoryArchivalURI:       "https://test.history",
						VisibilityArchivalURI:    "https://test.visibility",
						IsolationGroups:          types.IsolationGroupConfiguration{},
						AsyncWorkflowConfig:      types.AsyncWorkflowConfiguration{Enabled: false},
					},
					ReplicationConfig: &persistence.DomainReplicationConfig{
						ActiveClusterName: "active",
						Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: "active"}, {ClusterName: "standby"}},
					},
					ConfigVersion:   1,
					FailoverVersion: 123,
					IsGlobalDomain:  true,
					LastUpdatedTime: time.Now().Add(-24 * time.Hour).UnixNano(),
				}, nil)

				mockDomainManager.EXPECT().UpdateDomain(gomock.Any(), gomock.Any()).Return(nil)

				mockReplicator.EXPECT().HandleTransmissionTask(
					gomock.Any(),
					types.DomainOperationUpdate,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					true,
				).Return(fmt.Errorf("replication error"))
			},
			request: &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "global-test-domain",
				Configuration: &types.AsyncWorkflowConfiguration{
					Enabled: true,
				},
			},
			expectedErr: fmt.Errorf("replication error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockDomainManager := persistence.NewMockDomainManager(ctrl)
			mockReplicator := NewMockReplicator(ctrl)
			handler := newTestHandler(mockDomainManager, test.isPrimaryCluster, mockReplicator)
			test.setupMocks(mockDomainManager, mockReplicator)

			err := handler.UpdateAsyncWorkflowConfiguraton(context.Background(), *test.request)
			if test.expectedErr != nil {
				require.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

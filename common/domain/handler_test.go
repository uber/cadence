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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

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
func newTestHandler(domainManager persistence.DomainManager, primaryCluster bool) Handler {
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
	archivalMetadata := archiver.NewArchivalMetadata(mockDC, "Disabled", false, "Disabled", false, domainDefaults)

	testConfig := Config{
		MinRetentionDays:       dynamicconfig.GetIntPropertyFn(1),
		MaxRetentionDays:       dynamicconfig.GetIntPropertyFn(5),
		RequiredDomainDataKeys: nil,
		MaxBadBinaryCount:      nil,
		FailoverCoolDown:       nil,
	}

	return NewHandler(
		testConfig,
		log.NewNoop(),
		domainManager,
		cluster.GetTestClusterMetadata(primaryCluster),
		nil, // Replicator
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
		mockSetup        func(*persistence.MockDomainManager, *types.RegisterDomainRequest)
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
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
			mockSetup: func(mockDomainMgr *persistence.MockDomainManager, request *types.RegisterDomainRequest) {
				mockDomainMgr.EXPECT().GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: request.Name}).Return(nil, &types.EntityNotExistsError{})
			},
			wantErr:     true,
			expectedErr: &types.BadRequestError{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)

			mockDomainMgr := persistence.NewMockDomainManager(controller)
			handler := newTestHandler(mockDomainMgr, tc.isPrimaryCluster)

			tc.mockSetup(mockDomainMgr, tc.request)

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

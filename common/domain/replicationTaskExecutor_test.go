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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func TestDomainReplicationTaskExecutor_Execute(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mockDomainManager persistence.MockDomainManager)
		task      *types.DomainTaskAttributes
		wantErr   bool
		errType   interface{}
	}{
		{
			name: "Validate Domain Task - Empty Task",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				// No setup required as the task itself is nil, triggering the validation error
			},
			task:    nil,
			wantErr: true,
			errType: &types.BadRequestError{},
		},
		{
			name: "Handle Create Domain Task - Valid",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				mockDomainManager.EXPECT().
					CreateDomain(gomock.Any(), gomock.Any()).
					Return(&persistence.CreateDomainResponse{ID: "validDomainID"}, nil).
					Times(1)
			},
			task: &types.DomainTaskAttributes{
				DomainOperation: types.DomainOperationCreate.Ptr(),
				ID:              "validDomainID",
				Info: &types.DomainInfo{
					Name:        "validDomain",
					Status:      types.DomainStatusRegistered.Ptr(),
					Description: "A valid domain",
					OwnerEmail:  "owner@example.com",
					Data:        map[string]string{"k1": "v1"},
				},
				Config: &types.DomainConfiguration{
					WorkflowExecutionRetentionPeriodInDays: 7,
					EmitMetric:                             true,
					HistoryArchivalStatus:                  types.ArchivalStatusEnabled.Ptr(),
					HistoryArchivalURI:                     "test://history",
					VisibilityArchivalStatus:               types.ArchivalStatusEnabled.Ptr(),
					VisibilityArchivalURI:                  "test://visibility",
				},
				ReplicationConfig: &types.DomainReplicationConfiguration{
					ActiveClusterName: "activeClusterName",
					Clusters: []*types.ClusterReplicationConfiguration{
						{ClusterName: "activeClusterName"},
						{ClusterName: "standbyClusterName"},
					},
				},
				ConfigVersion:   1,
				FailoverVersion: 1,
			},
			wantErr: false,
		},
		{
			name: "Handle Create Domain Task - Name UUID Collision",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				// call to GetDomain simulates a name collision by returning a different domain ID
				mockDomainManager.EXPECT().
					GetDomain(gomock.Any(), &persistence.GetDomainRequest{Name: "collisionDomain"}).
					Return(&persistence.GetDomainResponse{Info: &persistence.DomainInfo{ID: uuid.New()}}, nil).
					Times(1)

				// Expect CreateDomain to be called, which should result in a collision error
				mockDomainManager.EXPECT().
					CreateDomain(gomock.Any(), gomock.Any()).
					Return(nil, ErrNameUUIDCollision).
					Times(1)
			},
			task: &types.DomainTaskAttributes{
				DomainOperation: types.DomainOperationCreate.Ptr(),
				ID:              uuid.New(),
				Info: &types.DomainInfo{
					Name:        "collisionDomain",
					Status:      types.DomainStatusRegistered.Ptr(),
					Description: "A domain with UUID collision",
					OwnerEmail:  "owner@example.com",
					Data:        map[string]string{"k1": "v1"},
				},
				Config: &types.DomainConfiguration{
					WorkflowExecutionRetentionPeriodInDays: 7,
					EmitMetric:                             true,
					HistoryArchivalStatus:                  types.ArchivalStatusEnabled.Ptr(),
					HistoryArchivalURI:                     "test://history",
					VisibilityArchivalStatus:               types.ArchivalStatusEnabled.Ptr(),
					VisibilityArchivalURI:                  "test://visibility",
				},
				ReplicationConfig: &types.DomainReplicationConfiguration{
					ActiveClusterName: "activeClusterName",
					Clusters: []*types.ClusterReplicationConfiguration{
						{ClusterName: "activeClusterName"},
						{ClusterName: "standbyClusterName"},
					},
				},
				ConfigVersion:   1,
				FailoverVersion: 1,
			},
			wantErr: true,
			errType: &types.BadRequestError{},
		},
		{
			name: "Handle Update Domain Task - Valid Update",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				mockDomainManager.EXPECT().
					GetMetadata(gomock.Any()).
					Return(&persistence.GetMetadataResponse{NotificationVersion: 123}, nil).
					Times(1)

				// Mock GetDomain to simulate domain fetch before update
				mockDomainManager.EXPECT().
					GetDomain(gomock.Any(), gomock.Any()).
					Return(&persistence.GetDomainResponse{
						Info:              &persistence.DomainInfo{ID: "existingDomainID", Name: "existingDomainName"},
						Config:            &persistence.DomainConfig{},
						ReplicationConfig: &persistence.DomainReplicationConfig{},
					}, nil).AnyTimes()

				// Mock UpdateDomain to simulate a successful domain update
				mockDomainManager.EXPECT().
					UpdateDomain(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			task: &types.DomainTaskAttributes{
				DomainOperation: types.DomainOperationUpdate.Ptr(),
				ID:              "existingDomainID",
				Info: &types.DomainInfo{
					Name:        "existingDomainName",
					Status:      types.DomainStatusRegistered.Ptr(),
					Description: "Updated description",
					OwnerEmail:  "updatedOwner@example.com",
					Data:        map[string]string{"updatedKey": "updatedValue"},
				},
				Config:            &types.DomainConfiguration{},
				ReplicationConfig: &types.DomainReplicationConfiguration{},
				ConfigVersion:     2,
				FailoverVersion:   100,
			},
			wantErr: false,
		},
		{
			name: "Handle Invalid Domain Operation",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				// No mock setup is required as the operation should not proceed to any database calls
			},
			task: &types.DomainTaskAttributes{
				// Set up a task without a valid DomainOperation or with an unrecognized operation
				DomainOperation: nil, // Assuming this would not match any specific case
				ID:              "invalidOperationDomainID",
				Info: &types.DomainInfo{
					Name:        "invalidOperationDomain",
					Status:      types.DomainStatusRegistered.Ptr(),
					Description: "Domain with invalid operation",
					OwnerEmail:  "owner@example.com",
					Data:        map[string]string{"k1": "v1"},
				},
				Config:            &types.DomainConfiguration{},
				ReplicationConfig: &types.DomainReplicationConfiguration{},
			},
			wantErr: true,
			errType: &types.BadRequestError{},
		},
		{
			name: "Handle Unsupported Domain Operation",
			setupMock: func(mockDomainManager persistence.MockDomainManager) {
				// No mock setup is needed as the operation should immediately return an error
			},
			task: &types.DomainTaskAttributes{
				DomainOperation: types.DomainOperation(999).Ptr(), // Assuming 999 is not a valid operation
				ID:              "someDomainID",
				Info: &types.DomainInfo{
					Name:        "someDomain",
					Status:      types.DomainStatusRegistered.Ptr(),
					Description: "A domain with an unsupported operation",
					OwnerEmail:  "email@example.com",
					Data:        map[string]string{"key": "value"},
				},
				Config: &types.DomainConfiguration{
					WorkflowExecutionRetentionPeriodInDays: 7,
					EmitMetric:                             true,
					HistoryArchivalStatus:                  types.ArchivalStatusEnabled.Ptr(),
					HistoryArchivalURI:                     "uri://history",
					VisibilityArchivalStatus:               types.ArchivalStatusEnabled.Ptr(),
					VisibilityArchivalURI:                  "uri://visibility",
				},
				ReplicationConfig: &types.DomainReplicationConfiguration{
					ActiveClusterName: "activeCluster",
					Clusters:          []*types.ClusterReplicationConfiguration{{ClusterName: "activeCluster"}},
				},
			},
			wantErr: true,
			errType: &types.BadRequestError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockDomainManager := persistence.NewMockDomainManager(ctrl)
			mockTimeSource := clock.NewRealTimeSource()
			mockLogger := log.NewNoop()

			executor := NewReplicationTaskExecutor(mockDomainManager, mockTimeSource, mockLogger).(*domainReplicationTaskExecutorImpl)
			tt.setupMock(*mockDomainManager)
			err := executor.Execute(tt.task)
			if tt.wantErr {
				require.Error(t, err)
				assert.IsType(t, tt.errType, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

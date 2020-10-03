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

package domain

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/resource"
)

type (
	domainReplicationTaskExecutorSuite struct {
		suite.Suite
		controller *gomock.Controller
		*require.Assertions

		mockMetadataMgr *mocks.MetadataManager

		domainReplicator *domainReplicationTaskExecutorImpl
	}
)

func TestDomainReplicationTaskExecutorSuite(t *testing.T) {
	s := new(domainReplicationTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *domainReplicationTaskExecutorSuite) SetupSuite() {
}

func (s *domainReplicationTaskExecutorSuite) TearDownSuite() {

}

func (s *domainReplicationTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())

	res := resource.NewTest(s.controller, metrics.PersistenceDomainReplicationQueueScope)
	s.mockMetadataMgr = res.MetadataMgr
	s.domainReplicator = NewReplicationTaskExecutor(res).(*domainReplicationTaskExecutorImpl)
}

func (s *domainReplicationTaskExecutorSuite) TearDownTest() {

}

func (s *domainReplicationTaskExecutorSuite) TestExecute_RegisterDomainTask_NameUUIDCollision() {
	operation := replicator.DomainOperationCreate
	id := uuid.New()
	name := "some random domain test name"
	status := shared.DomainStatusRegistered
	description := "some random test description"
	ownerEmail := "some random test owner"
	data := map[string]string{"k": "v"}
	retention := int32(10)
	emitMetric := true
	historyArchivalStatus := shared.ArchivalStatusEnabled
	historyArchivalURI := "some random history archival uri"
	visibilityArchivalStatus := shared.ArchivalStatusEnabled
	visibilityArchivalURI := "some random visibility archival uri"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(0)
	failoverVersion := int64(59)
	clusters := []*shared.ClusterReplicationConfiguration{
		{
			ClusterName: common.StringPtr(clusterActive),
		},
		{
			ClusterName: common.StringPtr(clusterStandby),
		},
	}

	task := &replicator.DomainTaskAttributes{
		DomainOperation: &operation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &status,
			Description: common.StringPtr(description),
			OwnerEmail:  common.StringPtr(ownerEmail),
			Data:        data,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(historyArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(historyArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(visibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(visibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		ConfigVersion:   common.Int64Ptr(configVersion),
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}

	s.mockMetadataMgr.On("CreateDomain", mock.Anything).Return(nil, errors.New("test")).Times(1)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(&persistence.GetDomainResponse{
		Info: &persistence.DomainInfo{
			ID: *task.ID,
		},
	}, nil).Times(1)

	task.ID = common.StringPtr(uuid.New())
	task.Info.Name = common.StringPtr(name)
	err := s.domainReplicator.Execute(task)
	s.NotNil(err)
	s.IsType(&shared.BadRequestError{}, err)
}

func (s *domainReplicationTaskExecutorSuite) TestExecute_RegisterDomainTask() {
	operation := replicator.DomainOperationCreate
	id := uuid.New()
	name := "some random domain test name"
	status := shared.DomainStatusRegistered
	description := "some random test description"
	ownerEmail := "some random test owner"
	data := map[string]string{"k": "v"}
	retention := int32(10)
	emitMetric := true
	historyArchivalStatus := shared.ArchivalStatusEnabled
	historyArchivalURI := "some random history archival uri"
	visibilityArchivalStatus := shared.ArchivalStatusEnabled
	visibilityArchivalURI := "some random visibility archival uri"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(0)
	failoverVersion := int64(59)
	clusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterStandby),
		},
	}

	task := &replicator.DomainTaskAttributes{
		DomainOperation: &operation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &status,
			Description: common.StringPtr(description),
			OwnerEmail:  common.StringPtr(ownerEmail),
			Data:        data,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(historyArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(historyArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(visibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(visibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		ConfigVersion:   common.Int64Ptr(configVersion),
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}

	s.mockMetadataMgr.On("CreateDomain", mock.Anything).Return(nil, nil).Times(1)
	// handle duplicated task
	err := s.domainReplicator.Execute(task)
	s.Nil(err)
}

func (s *domainReplicationTaskExecutorSuite) TestExecute_UpdateDomainTask_DomainNotExist() {
	operation := replicator.DomainOperationUpdate
	id := uuid.New()
	name := "some random domain test name"
	status := shared.DomainStatusRegistered
	description := "some random test description"
	ownerEmail := "some random test owner"
	retention := int32(10)
	emitMetric := true
	historyArchivalStatus := shared.ArchivalStatusEnabled
	historyArchivalURI := "some random history archival uri"
	visibilityArchivalStatus := shared.ArchivalStatusEnabled
	visibilityArchivalURI := "some random visibility archival uri"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(12)
	failoverVersion := int64(59)
	domainData := map[string]string{"k1": "v1", "k2": "v2"}
	clusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterStandby),
		},
	}

	updateTask := &replicator.DomainTaskAttributes{
		DomainOperation: &operation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &status,
			Description: common.StringPtr(description),
			OwnerEmail:  common.StringPtr(ownerEmail),
			Data:        domainData,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(historyArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(historyArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(visibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(visibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		ConfigVersion:   common.Int64Ptr(configVersion),
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}

	s.mockMetadataMgr.On("GetMetadata").Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(nil, &shared.EntityNotExistsError{Message: "No domain"}).Times(1)
	s.mockMetadataMgr.On("CreateDomain", mock.Anything).Return(nil, nil).Times(1)
	err := s.domainReplicator.Execute(updateTask)
	s.Nil(err)
}

func (s *domainReplicationTaskExecutorSuite) TestExecute_UpdateDomainTask_UpdateConfig_UpdateActiveCluster() {
	id := uuid.New()
	name := "some random domain test name"
	configVersion := int64(0)
	failoverVersion := int64(59)
	// success update case
	updateOperation := replicator.DomainOperationUpdate
	updateStatus := shared.DomainStatusDeprecated
	updateDescription := "other random domain test description"
	updateOwnerEmail := "other random domain test owner"
	updatedData := map[string]string{"k": "v1"}
	updateRetention := int32(122)
	updateEmitMetric := true
	updateHistoryArchivalStatus := shared.ArchivalStatusDisabled
	updateHistoryArchivalURI := "some updated history archival uri"
	updateVisibilityArchivalStatus := shared.ArchivalStatusDisabled
	updateVisibilityArchivalURI := "some updated visibility archival uri"
	updateClusterActive := "other random active cluster name"
	updateClusterStandby := "other random standby cluster name"
	updateConfigVersion := configVersion + 1
	updateFailoverVersion := failoverVersion + 1
	updateClusters := []*shared.ClusterReplicationConfiguration{
		{
			ClusterName: common.StringPtr(updateClusterActive),
		},
		{
			ClusterName: common.StringPtr(updateClusterStandby),
		},
	}
	updateTask := &replicator.DomainTaskAttributes{
		DomainOperation: &updateOperation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &updateStatus,
			Description: common.StringPtr(updateDescription),
			OwnerEmail:  common.StringPtr(updateOwnerEmail),
			Data:        updatedData,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(updateRetention),
			EmitMetric:                             common.BoolPtr(updateEmitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(updateHistoryArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(updateHistoryArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(updateVisibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(updateVisibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(updateClusterActive),
			Clusters:          updateClusters,
		},
		ConfigVersion:   common.Int64Ptr(updateConfigVersion),
		FailoverVersion: common.Int64Ptr(updateFailoverVersion),
	}

	s.mockMetadataMgr.On("GetMetadata").Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(&persistence.GetDomainResponse{
		Info:   nil,
		Config: nil,
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: updateClusterStandby,
			Clusters: []*persistence.ClusterReplicationConfig{
				{
					ClusterName: updateClusterActive,
				},
				{
					ClusterName: updateClusterStandby,
				},
			},
		},
		IsGlobalDomain:              true,
		ConfigVersion:               configVersion,
		FailoverVersion:             failoverVersion,
		FailoverNotificationVersion: 0,
		PreviousFailoverVersion:     0,
		FailoverEndTime:             nil,
		LastUpdatedTime:             0,
		NotificationVersion:         0,
	}, nil).Times(1)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything).Return(nil).Times(1)
	err := s.domainReplicator.Execute(updateTask)
	s.Nil(err)
}

func (s *domainReplicationTaskExecutorSuite) TestExecute_UpdateDomainTask_UpdateConfig_NoUpdateActiveCluster() {
	id := uuid.New()
	name := "some random domain test name"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(0)
	failoverVersion := int64(59)
	previousFailoverVersion := int64(55)

	// success update case
	updateOperation := replicator.DomainOperationUpdate
	updateStatus := shared.DomainStatusDeprecated
	updateDescription := "other random domain test description"
	updateOwnerEmail := "other random domain test owner"
	updateData := map[string]string{"k": "v2"}
	updateRetention := int32(122)
	updateEmitMetric := true
	updateHistoryArchivalStatus := shared.ArchivalStatusEnabled
	updateHistoryArchivalURI := "some updated history archival uri"
	updateVisibilityArchivalStatus := shared.ArchivalStatusEnabled
	updateVisibilityArchivalURI := "some updated visibility archival uri"
	updateClusterActive := "other random active cluster name"
	updateClusterStandby := "other random standby cluster name"
	updateConfigVersion := configVersion + 1
	updateFailoverVersion := failoverVersion - 1
	updateClusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterStandby),
		},
	}
	updateTask := &replicator.DomainTaskAttributes{
		DomainOperation: &updateOperation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &updateStatus,
			Description: common.StringPtr(updateDescription),
			OwnerEmail:  common.StringPtr(updateOwnerEmail),
			Data:        updateData,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(updateRetention),
			EmitMetric:                             common.BoolPtr(updateEmitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(updateHistoryArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(updateHistoryArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(updateVisibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(updateVisibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(updateClusterActive),
			Clusters:          updateClusters,
		},
		ConfigVersion:           common.Int64Ptr(updateConfigVersion),
		FailoverVersion:         common.Int64Ptr(updateFailoverVersion),
		PreviousFailoverVersion: common.Int64Ptr(previousFailoverVersion),
	}

	s.mockMetadataMgr.On("GetMetadata").Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(&persistence.GetDomainResponse{
		Info:   nil,
		Config: nil,
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: clusterStandby,
			Clusters: []*persistence.ClusterReplicationConfig{
				{
					ClusterName: clusterActive,
				},
				{
					ClusterName: clusterStandby,
				},
			},
		},
		IsGlobalDomain:              true,
		ConfigVersion:               configVersion,
		FailoverVersion:             failoverVersion,
		FailoverNotificationVersion: 0,
		PreviousFailoverVersion:     0,
		FailoverEndTime:             nil,
		LastUpdatedTime:             0,
		NotificationVersion:         0,
	}, nil).Times(1)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything).Return(nil).Times(1)
	err := s.domainReplicator.Execute(updateTask)
	s.Nil(err)

}

func (s *domainReplicationTaskExecutorSuite) TestExecute_UpdateDomainTask_NoUpdateConfig_UpdateActiveCluster() {
	id := uuid.New()
	name := "some random domain test name"
	historyArchivalStatus := shared.ArchivalStatusEnabled
	historyArchivalURI := "some random history archival uri"
	visibilityArchivalStatus := shared.ArchivalStatusEnabled
	visibilityArchivalURI := "some random visibility archival uri"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(0)
	failoverVersion := int64(59)

	// success update case
	updateOperation := replicator.DomainOperationUpdate
	updateStatus := shared.DomainStatusDeprecated
	updateDescription := "other random domain test description"
	updateOwnerEmail := "other random domain test owner"
	updatedData := map[string]string{"k": "v2"}
	updateRetention := int32(122)
	updateEmitMetric := true
	updateClusterActive := "other random active cluster name"
	updateClusterStandby := "other random standby cluster name"
	updateConfigVersion := configVersion - 1
	updateFailoverVersion := failoverVersion + 1
	updateClusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterStandby),
		},
	}
	updateTask := &replicator.DomainTaskAttributes{
		DomainOperation: &updateOperation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &updateStatus,
			Description: common.StringPtr(updateDescription),
			OwnerEmail:  common.StringPtr(updateOwnerEmail),
			Data:        updatedData,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(updateRetention),
			EmitMetric:                             common.BoolPtr(updateEmitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(historyArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(historyArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(visibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(visibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(updateClusterActive),
			Clusters:          updateClusters,
		},
		ConfigVersion:           common.Int64Ptr(updateConfigVersion),
		FailoverVersion:         common.Int64Ptr(updateFailoverVersion),
		PreviousFailoverVersion: common.Int64Ptr(failoverVersion),
	}

	s.mockMetadataMgr.On("GetMetadata").Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(&persistence.GetDomainResponse{
		Info:   nil,
		Config: nil,
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: clusterStandby,
			Clusters: []*persistence.ClusterReplicationConfig{
				{
					ClusterName: clusterActive,
				},
				{
					ClusterName: clusterStandby,
				},
			},
		},
		IsGlobalDomain:              true,
		ConfigVersion:               configVersion,
		FailoverVersion:             failoverVersion,
		FailoverNotificationVersion: 0,
		PreviousFailoverVersion:     0,
		FailoverEndTime:             nil,
		LastUpdatedTime:             0,
		NotificationVersion:         0,
	}, nil).Times(1)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything).Return(nil).Times(1)
	err := s.domainReplicator.Execute(updateTask)
	s.Nil(err)
}

func (s *domainReplicationTaskExecutorSuite) TestExecute_UpdateDomainTask_NoUpdateConfig_NoUpdateActiveCluster() {
	id := uuid.New()
	name := "some random domain test name"
	historyArchivalStatus := shared.ArchivalStatusEnabled
	historyArchivalURI := "some random history archival uri"
	visibilityArchivalStatus := shared.ArchivalStatusEnabled
	visibilityArchivalURI := "some random visibility archival uri"
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	configVersion := int64(0)
	failoverVersion := int64(59)
	previousFailoverVersion := int64(55)

	// success update case
	updateOperation := replicator.DomainOperationUpdate
	updateStatus := shared.DomainStatusDeprecated
	updateDescription := "other random domain test description"
	updateOwnerEmail := "other random domain test owner"
	updatedData := map[string]string{"k": "v2"}
	updateRetention := int32(122)
	updateEmitMetric := true
	updateClusterActive := "other random active cluster name"
	updateClusterStandby := "other random standby cluster name"
	updateConfigVersion := configVersion - 1
	updateFailoverVersion := failoverVersion - 1
	updateClusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(updateClusterStandby),
		},
	}
	updateTask := &replicator.DomainTaskAttributes{
		DomainOperation: &updateOperation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &updateStatus,
			Description: common.StringPtr(updateDescription),
			OwnerEmail:  common.StringPtr(updateOwnerEmail),
			Data:        updatedData,
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(updateRetention),
			EmitMetric:                             common.BoolPtr(updateEmitMetric),
			HistoryArchivalStatus:                  common.ArchivalStatusPtr(historyArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(historyArchivalURI),
			VisibilityArchivalStatus:               common.ArchivalStatusPtr(visibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(visibilityArchivalURI),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(updateClusterActive),
			Clusters:          updateClusters,
		},
		ConfigVersion:           common.Int64Ptr(updateConfigVersion),
		FailoverVersion:         common.Int64Ptr(updateFailoverVersion),
		PreviousFailoverVersion: common.Int64Ptr(previousFailoverVersion),
	}

	s.mockMetadataMgr.On("GetMetadata").Return(&persistence.GetMetadataResponse{NotificationVersion: 1}, nil)
	s.mockMetadataMgr.On("GetDomain", mock.Anything).Return(&persistence.GetDomainResponse{
		Info:   nil,
		Config: nil,
		ReplicationConfig: &persistence.DomainReplicationConfig{
			ActiveClusterName: clusterStandby,
			Clusters: []*persistence.ClusterReplicationConfig{
				{
					ClusterName: clusterActive,
				},
				{
					ClusterName: clusterStandby,
				},
			},
		},
		IsGlobalDomain:              true,
		ConfigVersion:               configVersion,
		FailoverVersion:             failoverVersion,
		FailoverNotificationVersion: 0,
		PreviousFailoverVersion:     0,
		FailoverEndTime:             nil,
		LastUpdatedTime:             0,
		NotificationVersion:         0,
	}, nil).Times(1)
	s.mockMetadataMgr.On("UpdateDomain", mock.Anything).Return(nil).Times(1)
	err := s.domainReplicator.Execute(updateTask)
	s.Nil(err)
}

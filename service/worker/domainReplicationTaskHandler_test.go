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

package worker

import (
	"log"
	"os"
	"testing"

	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	domainReplicatorSuite struct {
		suite.Suite
		persistence.TestBase
		domainReplicator *domainReplicatorImpl
	}
)

func TestDomainReplicatorSuite(t *testing.T) {
	s := new(domainReplicatorSuite)
	suite.Run(t, s)
}

func (s *domainReplicatorSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
	s.SetupWorkflowStore()
}

func (s *domainReplicatorSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

func (s *domainReplicatorSuite) SetupTest() {
	s.domainReplicator = newDomainReplicator(
		s.MetadataManager,
		bark.NewLoggerFromLogrus(logrus.New()),
	).(*domainReplicatorImpl)
}

func (s *domainReplicatorSuite) TearDownTest() {
}

func (s *domainReplicatorSuite) TearReplicateRegisterDomainTask() {
	operation := replicator.DomainOperationCreate
	id := uuid.New()
	name := "some random domain test name"
	status := shared.DomainStatusRegistered
	description := "some random test description"
	ownerEmail := "some random test owner"
	retention := int32(10)
	emitMetric := true
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
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
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}

	err := s.domainReplicator.HandleReceiveTask(task)
	s.Nil(err)

	resp, err := s.MetadataManager.GetDomain(&persistence.GetDomainRequest{ID: id})
	s.Nil(err)
	s.NotNil(resp)
	s.Equal(id, resp.Info.ID)
	s.Equal(name, resp.Info.Name)
	s.Equal(persistence.DomainStatusRegistered, resp.Info.Status)
	s.Equal(description, resp.Info.Description)
	s.Equal(ownerEmail, resp.Info.OwnerEmail)
	s.Equal(retention, resp.Config.Retention)
	s.Equal(emitMetric, resp.Config.EmitMetric)
	s.Equal(clusterActive, resp.ReplicationConfig.ActiveClusterName)
	s.Equal(failoverVersion, resp.ReplicationConfig.FailoverVersion)
	s.Equal(s.domainReplicator.convertClusterReplicationConfig(clusters), resp.ReplicationConfig.Clusters)
	s.Equal(int64(0), resp.Version)
}

func (s *domainReplicatorSuite) TearReplicateUpdateDomainTask() {
	operation := replicator.DomainOperationCreate
	id := uuid.New()
	name := "some random domain test name"
	status := shared.DomainStatusRegistered
	description := "some random test description"
	ownerEmail := "some random test owner"
	retention := int32(10)
	emitMetric := true
	clusterActive := "some random active cluster name"
	clusterStandby := "some random standby cluster name"
	failoverVersion := int64(59)
	clusters := []*shared.ClusterReplicationConfiguration{
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterActive),
		},
		&shared.ClusterReplicationConfiguration{
			ClusterName: common.StringPtr(clusterStandby),
		},
	}

	createTask := &replicator.DomainTaskAttributes{
		DomainOperation: &operation,
		ID:              common.StringPtr(id),
		Info: &shared.DomainInfo{
			Name:        common.StringPtr(name),
			Status:      &status,
			Description: common.StringPtr(description),
			OwnerEmail:  common.StringPtr(ownerEmail),
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}

	err := s.domainReplicator.HandleReceiveTask(createTask)
	s.Nil(err)

	// success update case
	operation = replicator.DomainOperationUpdate
	status = shared.DomainStatusDeprecated
	description = "other random domain test description"
	ownerEmail = "other random domain test owner"
	retention = int32(122)
	emitMetric = true
	clusterActive = "other random active cluster name"
	clusterStandby = "other random standby cluster name"
	failoverVersion = int64(100)
	clusters = []*shared.ClusterReplicationConfiguration{
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
		},
		Config: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retention),
			EmitMetric:                             common.BoolPtr(emitMetric),
		},
		ReplicationConfig: &shared.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(clusterActive),
			Clusters:          clusters,
		},
		FailoverVersion: common.Int64Ptr(failoverVersion),
	}
	err = s.domainReplicator.HandleReceiveTask(updateTask)
	s.Nil(err)
	resp, err := s.MetadataManager.GetDomain(&persistence.GetDomainRequest{Name: name})
	s.Nil(err)
	s.NotNil(resp)
	s.Equal(id, resp.Info.ID)
	s.Equal(name, resp.Info.Name)
	s.Equal(persistence.DomainStatusRegistered, resp.Info.Status)
	s.Equal(description, resp.Info.Description)
	s.Equal(ownerEmail, resp.Info.OwnerEmail)
	s.Equal(retention, resp.Config.Retention)
	s.Equal(emitMetric, resp.Config.EmitMetric)
	s.Equal(clusterActive, resp.ReplicationConfig.ActiveClusterName)
	s.Equal(failoverVersion, resp.ReplicationConfig.FailoverVersion)
	s.Equal(s.domainReplicator.convertClusterReplicationConfig(clusters), resp.ReplicationConfig.Clusters)
	s.Equal(int64(1), resp.Version)

	// ignored update case, since failover version is smaller
	updateTask.FailoverVersion = common.Int64Ptr(failoverVersion - 1)
	err = s.domainReplicator.HandleReceiveTask(updateTask)
	s.NotNil(err)
	resp, err = s.MetadataManager.GetDomain(&persistence.GetDomainRequest{Name: name})
	s.Nil(err)
	s.NotNil(resp)
	s.Equal(id, resp.Info.ID)
	s.Equal(name, resp.Info.Name)
	s.Equal(persistence.DomainStatusRegistered, resp.Info.Status)
	s.Equal(description, resp.Info.Description)
	s.Equal(ownerEmail, resp.Info.OwnerEmail)
	s.Equal(retention, resp.Config.Retention)
	s.Equal(emitMetric, resp.Config.EmitMetric)
	s.Equal(clusterActive, resp.ReplicationConfig.ActiveClusterName)
	s.Equal(failoverVersion, resp.ReplicationConfig.FailoverVersion)
	s.Equal(s.domainReplicator.convertClusterReplicationConfig(clusters), resp.ReplicationConfig.Clusters)
	s.Equal(int64(1), resp.Version)
}

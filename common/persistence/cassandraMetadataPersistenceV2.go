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

package persistence

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/uber-common/bark"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const constDomainPartition = 0
const initialFailoverGlobalNotificationVersion int64 = 0
const domainMetadataRecordName = "cadence-domain-metadata"

const (
	templateCreateDomainByNameQueryWithinBatchV2 = `INSERT INTO domains_by_name_v2 (` +
		`domains_partition, name, domain, config, replication_config, is_global_domain, config_version, failover_version, failover_global_notification_version, global_notification_version) ` +
		`VALUES(?, ?, ` + templateDomainInfoType + `, ` + templateDomainConfigType + `, ` + templateDomainReplicationConfigType + `, ?, ?, ?, ?, ?) IF NOT EXISTS`

	templateGetDomainByNameQueryV2 = `SELECT domain.id, domain.name, domain.status, domain.description, ` +
		`domain.owner_email, config.retention, config.emit_metric, ` +
		`replication_config.active_cluster_name, replication_config.clusters, ` +
		`is_global_domain, ` +
		`config_version, ` +
		`failover_version, ` +
		`failover_global_notification_version, ` +
		`global_notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateUpdateDomainByNameQueryWithinBatchV2 = `UPDATE domains_by_name_v2 ` +
		`SET domain = ` + templateDomainInfoType + `, ` +
		`config = ` + templateDomainConfigType + `, ` +
		`replication_config = ` + templateDomainReplicationConfigType + `, ` +
		`config_version = ? ,` +
		`failover_version = ? ,` +
		`failover_global_notification_version = ? , ` +
		`global_notification_version = ? ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateGetMetadataQueryV2 = `SELECT global_notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ? `

	templateUpdateMetadataQueryWithinBatchV2 = `UPDATE domains_by_name_v2 ` +
		`SET global_notification_version = ? ` +
		`WHERE domains_partition = ? ` +
		`and name = ? ` +
		`IF global_notification_version = ? `

	templateDeleteDomainByNameQueryV2 = `DELETE FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateListDomainQueryV2 = `SELECT name, domain.id, domain.name, domain.status, domain.description, ` +
		`domain.owner_email, config.retention, config.emit_metric, ` +
		`replication_config.active_cluster_name, replication_config.clusters, ` +
		`is_global_domain, ` +
		`config_version, ` +
		`failover_version, ` +
		`failover_global_notification_version, ` +
		`global_notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? `
)

type (
	cassandraMetadataPersistencev2 struct {
		session            *gocql.Session
		currentClusterName string
		logger             bark.Logger
	}
)

// NewCassandraMetadataPersistenceV2 is used to create an instance of HistoryManager implementation
func NewCassandraMetadataPersistenceV2(hosts string, port int, user, password, dc string, keyspace string,
	currentClusterName string, logger bark.Logger) (MetadataManagerV2, error) {
	cluster := common.NewCassandraCluster(hosts, port, user, password, dc)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = cassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = defaultSessionTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &cassandraMetadataPersistencev2{
		session:            session,
		currentClusterName: currentClusterName,
		logger:             logger,
	}, nil
}

// Close releases the resources held by this object
func (m *cassandraMetadataPersistencev2) Close() {
	if m.session != nil {
		m.session.Close()
	}
}

// Cassandra does not support conditional updates across multiple tables.  For this reason we have to first insert into
// 'Domains' table and then do a conditional insert into domains_by_name table.  If the conditional write fails we
// delete the orphaned entry from domains table.  There is a chance delete entry could fail and we never delete the
// orphaned entry from domains table.  We might need a background job to delete those orphaned record.
func (m *cassandraMetadataPersistencev2) CreateDomainV2(request *CreateDomainRequest) (*CreateDomainResponse, error) {
	query := m.session.Query(templateCreateDomainQuery, request.Info.ID, request.Info.Name)
	applied, err := query.ScanCAS()
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateDomain operation failed. Inserting into domains table. Error: %v", err),
		}
	}
	if !applied {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateDomain operation failed because of uuid collision."),
		}
	}

	globalNotificationVersion, err := m.GetMetadataV2()
	if err != nil {
		return nil, err
	}

	batch := m.session.NewBatch(gocql.LoggedBatch)
	batch.Query(templateCreateDomainByNameQueryWithinBatchV2,
		constDomainPartition,
		request.Info.Name,
		request.Info.ID,
		request.Info.Name,
		request.Info.Status,
		request.Info.Description,
		request.Info.OwnerEmail,
		request.Config.Retention,
		request.Config.EmitMetric,
		request.ReplicationConfig.ActiveClusterName,
		serializeClusterConfigs(request.ReplicationConfig.Clusters),
		request.IsGlobalDomain,
		request.ConfigVersion,
		request.FailoverVersion,
		initialFailoverGlobalNotificationVersion,
		globalNotificationVersion,
	)
	m.updateMetadataBatchV2(batch, globalNotificationVersion)

	previous := make(map[string]interface{})
	applied, iter, err := m.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateDomain operation failed. Inserting into domains_by_name_v2 table. Error: %v", err),
		}
	}

	if !applied {
		// Domain already exist.  Delete orphan domain record before returning back to user
		if errDelete := m.session.Query(templateDeleteDomainQuery, request.Info.ID).Exec(); errDelete != nil {
			m.logger.Warnf("Unable to delete orphan domain record. Error: %v", errDelete)
		}

		if domain, ok := previous["domain"].(map[string]interface{}); ok {
			msg := fmt.Sprintf("Domain already exists.  DomainId: %v", domain["id"])
			return nil, &workflow.DomainAlreadyExistsError{
				Message: msg,
			}
		}

		return nil, &workflow.DomainAlreadyExistsError{
			Message: fmt.Sprintf("CreateDomain operation failed because of conditional failure."),
		}
	}

	return &CreateDomainResponse{ID: request.Info.ID}, nil
}

func (m *cassandraMetadataPersistencev2) UpdateDomainV2(request *UpdateDomainRequest) error {
	batch := m.session.NewBatch(gocql.LoggedBatch)
	batch.Query(templateUpdateDomainByNameQueryWithinBatchV2,
		request.Info.ID,
		request.Info.Name,
		request.Info.Status,
		request.Info.Description,
		request.Info.OwnerEmail,
		request.Config.Retention,
		request.Config.EmitMetric,
		request.ReplicationConfig.ActiveClusterName,
		serializeClusterConfigs(request.ReplicationConfig.Clusters),
		request.ConfigVersion,
		request.FailoverVersion,
		request.FailoverGlobalNotificationVersion,
		request.GlobalNotificationVersion,
		constDomainPartition,
		request.Info.Name,
	)
	m.updateMetadataBatchV2(batch, request.GlobalNotificationVersion)

	previous := make(map[string]interface{})
	applied, iter, err := m.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateDomain operation failed. Error: %v", err),
		}
	}
	if !applied {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateDomain operation failed because of conditional failure."),
		}
	}

	return nil
}

func (m *cassandraMetadataPersistencev2) GetDomainV2(request *GetDomainRequest) (*GetDomainResponse, error) {
	var query *gocql.Query
	var err error
	info := &DomainInfo{}
	config := &DomainConfig{}
	replicationConfig := &DomainReplicationConfig{}
	var replicationClusters []map[string]interface{}
	var failoverGlobalNotificationVersion int64
	var globalNotificationVersion int64
	var failoverVersion int64
	var configVersion int64
	var isGlobalDomain bool

	if len(request.ID) > 0 && len(request.Name) > 0 {
		return nil, &workflow.BadRequestError{
			Message: "GetDomain operation failed.  Both ID and Name specified in request.",
		}
	} else if len(request.ID) == 0 && len(request.Name) == 0 {
		return nil, &workflow.BadRequestError{
			Message: "GetDomain operation failed.  Both ID and Name are empty.",
		}
	}

	handleError := func(name, ID string, err error) error {
		identity := name
		if len(ID) > 0 {
			identity = ID
		}
		if err == gocql.ErrNotFound {
			return &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Domain %s does not exist.", identity),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetDomain operation failed. Error %v", err),
		}
	}

	domainName := request.Name
	if len(request.ID) > 0 {
		query = m.session.Query(templateGetDomainQuery, request.ID)
		err = query.Scan(&domainName)
		if err != nil {
			return nil, handleError(request.Name, request.ID, err)
		}
	}

	query = m.session.Query(templateGetDomainByNameQueryV2, constDomainPartition, domainName)
	err = query.Scan(
		&info.ID,
		&info.Name,
		&info.Status,
		&info.Description,
		&info.OwnerEmail,
		&config.Retention,
		&config.EmitMetric,
		&replicationConfig.ActiveClusterName,
		&replicationClusters,
		&isGlobalDomain,
		&configVersion,
		&failoverVersion,
		&failoverGlobalNotificationVersion,
		&globalNotificationVersion,
	)

	if err != nil {
		return nil, handleError(request.Name, request.ID, err)
	}

	replicationConfig.ActiveClusterName = GetOrUseDefaultActiveCluster(m.currentClusterName, replicationConfig.ActiveClusterName)
	replicationConfig.Clusters = deserializeClusterConfigs(replicationClusters)
	replicationConfig.Clusters = GetOrUseDefaultClusters(m.currentClusterName, replicationConfig.Clusters)

	return &GetDomainResponse{
		Info:                              info,
		Config:                            config,
		ReplicationConfig:                 replicationConfig,
		IsGlobalDomain:                    isGlobalDomain,
		ConfigVersion:                     configVersion,
		FailoverVersion:                   failoverVersion,
		FailoverGlobalNotificationVersion: failoverGlobalNotificationVersion,
		GlobalNotificationVersion:         globalNotificationVersion,
	}, nil
}

func (m *cassandraMetadataPersistencev2) ListDomainV2(request *ListDomainRequest) (*ListDomainResponse, error) {
	var query *gocql.Query

	query = m.session.Query(templateListDomainQueryV2, constDomainPartition)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "ListDomains operation failed.  Not able to create query iterator.",
		}
	}

	var name string
	domain := &GetDomainResponse{
		Info:              &DomainInfo{},
		Config:            &DomainConfig{},
		ReplicationConfig: &DomainReplicationConfig{},
	}
	var replicationClusters []map[string]interface{}
	response := &ListDomainResponse{}
	for iter.Scan(
		&name,
		&domain.Info.ID, &domain.Info.Name, &domain.Info.Status, &domain.Info.Description, &domain.Info.OwnerEmail,
		&domain.Config.Retention, &domain.Config.EmitMetric,
		&domain.ReplicationConfig.ActiveClusterName, &replicationClusters,
		&domain.IsGlobalDomain, &domain.ConfigVersion, &domain.FailoverVersion,
		&domain.FailoverGlobalNotificationVersion, &domain.GlobalNotificationVersion,
	) {
		if name != domainMetadataRecordName {
			// do not inlcude the metadata record
			domain.ReplicationConfig.ActiveClusterName = GetOrUseDefaultActiveCluster(m.currentClusterName, domain.ReplicationConfig.ActiveClusterName)
			domain.ReplicationConfig.Clusters = deserializeClusterConfigs(replicationClusters)
			domain.ReplicationConfig.Clusters = GetOrUseDefaultClusters(m.currentClusterName, domain.ReplicationConfig.Clusters)
			response.Domains = append(response.Domains, domain)
		}
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListDomains operation failed. Error: %v", err),
		}
	}

	return response, nil
}

func (m *cassandraMetadataPersistencev2) DeleteDomainV2(request *DeleteDomainRequest) error {
	var name string
	query := m.session.Query(templateGetDomainQuery, request.ID)
	err := query.Scan(&name)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil
		}
		return err
	}

	return m.deleteDomainV2(name, request.ID)
}

func (m *cassandraMetadataPersistencev2) DeleteDomainByNameV2(request *DeleteDomainByNameRequest) error {
	var ID string
	query := m.session.Query(templateGetDomainByNameQueryV2, constDomainPartition, request.Name)
	err := query.Scan(&ID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil
		}
		return err
	}
	return m.deleteDomainV2(request.Name, ID)
}

func (m *cassandraMetadataPersistencev2) GetMetadataV2() (int64, error) {
	var globalNotificationVersion int64
	query := m.session.Query(templateGetMetadataQueryV2, constDomainPartition, domainMetadataRecordName)
	err := query.Scan(&globalNotificationVersion)
	if err != nil {
		if err == gocql.ErrNotFound {
			// this error can be thrown in the very begining,
			// i.e. when domains_by_name_v2 is initialized
			m.logger.Warn("Cannot find domain metadata record.")
			return 0, nil
		}
		return 0, err
	}
	return globalNotificationVersion, nil
}

func (m *cassandraMetadataPersistencev2) updateMetadataBatchV2(batch *gocql.Batch, globalNotificationVersion int64) {
	var nextVersion int64 = 1
	var currentVersion *int64
	if globalNotificationVersion > 0 {
		nextVersion = globalNotificationVersion + 1
		currentVersion = &globalNotificationVersion
	}
	batch.Query(templateUpdateMetadataQueryWithinBatchV2,
		nextVersion,
		constDomainPartition,
		domainMetadataRecordName,
		currentVersion,
	)
}

func (m *cassandraMetadataPersistencev2) deleteDomainV2(name, ID string) error {
	query := m.session.Query(templateDeleteDomainByNameQueryV2, constDomainPartition, name)
	if err := query.Exec(); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("DeleteDomainByName operation failed. Error %v", err),
		}
	}

	query = m.session.Query(templateDeleteDomainQuery, ID)
	if err := query.Exec(); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("DeleteDomain operation failed. Error %v", err),
		}
	}

	return nil
}

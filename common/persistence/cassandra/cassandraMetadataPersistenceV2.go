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

package cassandra

import (
	"context"
	"fmt"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"

	"github.com/gocql/gocql"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/service/config"
)

const (
	constDomainPartition     = 0
	domainMetadataRecordName = "cadence-domain-metadata"
	emptyFailoverEndTime     = int64(0)
)

const (
	templateDomainInfoType = `{` +
		`id: ?, ` +
		`name: ?, ` +
		`status: ?, ` +
		`description: ?, ` +
		`owner_email: ?, ` +
		`data: ? ` +
		`}`

	templateDomainConfigType = `{` +
		`retention: ?, ` +
		`emit_metric: ?, ` +
		`archival_bucket: ?, ` +
		`archival_status: ?,` +
		`history_archival_status: ?, ` +
		`history_archival_uri: ?, ` +
		`visibility_archival_status: ?, ` +
		`visibility_archival_uri: ?, ` +
		`bad_binaries: ?,` +
		`bad_binaries_encoding: ?` +
		`}`

	templateDomainReplicationConfigType = `{` +
		`active_cluster_name: ?, ` +
		`clusters: ? ` +
		`}`

	templateCreateDomainQuery = `INSERT INTO domains (` +
		`id, domain) ` +
		`VALUES(?, {name: ?}) IF NOT EXISTS`

	templateGetDomainQuery = `SELECT domain.name ` +
		`FROM domains ` +
		`WHERE id = ?`

	templateDeleteDomainQuery = `DELETE FROM domains ` +
		`WHERE id = ?`

	templateCreateDomainByNameQueryWithinBatchV2 = `INSERT INTO domains_by_name_v2 (` +
		`domains_partition, name, domain, config, replication_config, is_global_domain, config_version, failover_version, failover_notification_version, previous_failover_version, failover_end_time, notification_version) ` +
		`VALUES(?, ?, ` + templateDomainInfoType + `, ` + templateDomainConfigType + `, ` + templateDomainReplicationConfigType + `, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS`

	templateGetDomainByNameQueryV2 = `SELECT domain.id, domain.name, domain.status, domain.description, ` +
		`domain.owner_email, domain.data, config.retention, config.emit_metric, ` +
		`config.archival_bucket, config.archival_status, ` +
		`config.history_archival_status, config.history_archival_uri, ` +
		`config.visibility_archival_status, config.visibility_archival_uri, ` +
		`config.bad_binaries, config.bad_binaries_encoding, ` +
		`replication_config.active_cluster_name, replication_config.clusters, ` +
		`is_global_domain, ` +
		`config_version, ` +
		`failover_version, ` +
		`failover_notification_version, ` +
		`previous_failover_version, ` +
		`failover_end_time, ` +
		`notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateUpdateDomainByNameQueryWithinBatchV2 = `UPDATE domains_by_name_v2 ` +
		`SET domain = ` + templateDomainInfoType + `, ` +
		`config = ` + templateDomainConfigType + `, ` +
		`replication_config = ` + templateDomainReplicationConfigType + `, ` +
		`config_version = ? ,` +
		`failover_version = ? ,` +
		`failover_notification_version = ? , ` +
		`previous_failover_version = ? , ` +
		`failover_end_time = ?,` +
		`notification_version = ? ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateGetMetadataQueryV2 = `SELECT notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ? `

	templateUpdateMetadataQueryWithinBatchV2 = `UPDATE domains_by_name_v2 ` +
		`SET notification_version = ? ` +
		`WHERE domains_partition = ? ` +
		`and name = ? ` +
		`IF notification_version = ? `

	templateDeleteDomainByNameQueryV2 = `DELETE FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? ` +
		`and name = ?`

	templateListDomainQueryV2 = `SELECT name, domain.id, domain.name, domain.status, domain.description, ` +
		`domain.owner_email, domain.data, config.retention, config.emit_metric, ` +
		`config.archival_bucket, config.archival_status, ` +
		`config.history_archival_status, config.history_archival_uri, ` +
		`config.visibility_archival_status, config.visibility_archival_uri, ` +
		`config.bad_binaries, config.bad_binaries_encoding, ` +
		`replication_config.active_cluster_name, replication_config.clusters, ` +
		`is_global_domain, ` +
		`config_version, ` +
		`failover_version, ` +
		`failover_notification_version, ` +
		`previous_failover_version, ` +
		`failover_end_time, ` +
		`notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? `
)

type (
	nosqlDomainManager struct {
		nosqlManager
		currentClusterName string
	}
)

// newMetadataPersistenceV2 is used to create an instance of HistoryManager implementation
func newMetadataPersistenceV2(cfg config.Cassandra, currentClusterName string, logger log.Logger) (p.MetadataStore, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlDomainManager{
		nosqlManager: nosqlManager{
			db:     db,
			logger: logger,
		},
		currentClusterName: currentClusterName,
	}, nil
}

// CreateDomain create a domain
// Cassandra does not support conditional updates across multiple tables.  For this reason we have to first insert into
// 'Domains' table and then do a conditional insert into domains_by_name table.  If the conditional write fails we
// delete the orphaned entry from domains table.  There is a chance delete entry could fail and we never delete the
// orphaned entry from domains table.  We might need a background job to delete those orphaned record.
func (m *nosqlDomainManager) CreateDomain(
	ctx context.Context,
	request *p.InternalCreateDomainRequest,
) (*p.CreateDomainResponse, error) {
	row := &nosqlplugin.DomainRow{
		Info:                        request.Info,
		Config:                      request.Config,
		ReplicationConfig:           request.ReplicationConfig,
		ConfigVersion:               request.ConfigVersion,
		FailoverVersion:             request.FailoverVersion,
		FailoverNotificationVersion: p.InitialFailoverNotificationVersion,
		PreviousFailoverVersion:     common.InitialPreviousFailoverVersion,
		FailoverEndTime:             emptyFailoverEndTime,
		IsGlobalDomain:              request.IsGlobalDomain,
	}

	err := m.db.InsertDomain(ctx, row)

	if err != nil {
		if m.db.IsConditionFailedError(err) {
			return nil, &workflow.DomainAlreadyExistsError{
				Message: fmt.Sprintf("CreateDomain operation failed because of conditional failure, %v", err),
			}
		}
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateDomain operation failed. Inserting into domains table. Error: %v", err),
		}
	}

	return &p.CreateDomainResponse{ID: request.Info.ID}, nil
}

func (m *nosqlDomainManager) UpdateDomain(
	ctx context.Context,
	request *p.InternalUpdateDomainRequest,
) error {
	failoverEndTime := emptyFailoverEndTime
	if request.FailoverEndTime != nil {
		failoverEndTime = *request.FailoverEndTime
	}

	row := &nosqlplugin.DomainRow{
		Info:                        request.Info,
		Config:                      request.Config,
		ReplicationConfig:           request.ReplicationConfig,
		ConfigVersion:               request.ConfigVersion,
		FailoverVersion:             request.FailoverVersion,
		FailoverNotificationVersion: request.FailoverNotificationVersion,
		PreviousFailoverVersion:     request.PreviousFailoverVersion,
		FailoverEndTime:             failoverEndTime,
		NotificationVersion:         request.NotificationVersion,
	}

	err := m.db.UpdateDomain(ctx, row)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateDomain operation failed. Error: %v", err),
		}
	}

	return nil
}

func (m *nosqlDomainManager) GetDomain(
	ctx context.Context,
	request *p.GetDomainRequest,
) (*p.InternalGetDomainResponse, error) {
	if len(request.ID) > 0 && len(request.Name) > 0 {
		return nil, &workflow.BadRequestError{
			Message: "GetDomain operation failed.  Both ID and Name specified in request.",
		}
	} else if len(request.ID) == 0 && len(request.Name) == 0 {
		return nil, &workflow.BadRequestError{
			Message: "GetDomain operation failed.  Both ID and Name are empty.",
		}
	}
	var domainName *string
	var domainID *string
	if len(request.ID) > 0 {
		domainID = common.StringPtr(request.ID)
	} else {
		domainName = common.StringPtr(request.Name)
	}

	handleError := func(name, ID string, err error) error {
		identity := name
		if len(ID) > 0 {
			identity = ID
		}
		if m.db.IsNotFoundError(err) {
			return &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Domain %s does not exist.", identity),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetDomain operation failed. Error %v", err),
		}
	}

	row, err := m.db.SelectDomain(ctx, domainID, domainName)

	if err != nil {
		return nil, handleError(request.Name, request.ID, err)
	}

	if row.Info.Data == nil {
		row.Info.Data = map[string]string{}
	}
	row.ReplicationConfig.ActiveClusterName = p.GetOrUseDefaultActiveCluster(m.currentClusterName, row.ReplicationConfig.ActiveClusterName)
	row.ReplicationConfig.Clusters = p.GetOrUseDefaultClusters(m.currentClusterName, row.ReplicationConfig.Clusters)

	// Note: to make it nullable
	var responseFailoverEndTime *int64
	if row.FailoverEndTime > emptyFailoverEndTime {
		domainFailoverEndTime := row.FailoverEndTime
		responseFailoverEndTime = common.Int64Ptr(domainFailoverEndTime)
	}

	return &p.InternalGetDomainResponse{
		Info:                        row.Info,
		Config:                      row.Config,
		ReplicationConfig:           row.ReplicationConfig,
		IsGlobalDomain:              row.IsGlobalDomain,
		ConfigVersion:               row.ConfigVersion,
		FailoverVersion:             row.FailoverVersion,
		FailoverNotificationVersion: row.FailoverNotificationVersion,
		PreviousFailoverVersion:     row.PreviousFailoverVersion,
		FailoverEndTime:             responseFailoverEndTime,
		NotificationVersion:         row.NotificationVersion,
	}, nil
}

func (m *nosqlDomainManager) ListDomains(
	ctx context.Context,
	request *p.ListDomainsRequest,
) (*p.InternalListDomainsResponse, error) {
	rows, nextPageToken, err := m.db.SelectAllDomains(ctx, request.PageSize, request.NextPageToken)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ListDomains operation failed. Error: %v", err),
		}
	}
	var domains []*p.InternalGetDomainResponse
	for _, row := range rows {
		row.ReplicationConfig.ActiveClusterName = p.GetOrUseDefaultActiveCluster(m.currentClusterName, row.ReplicationConfig.ActiveClusterName)
		row.ReplicationConfig.Clusters = p.GetOrUseDefaultClusters(m.currentClusterName, row.ReplicationConfig.Clusters)

		// Note: to make it nullable
		var domainFailoverEndTime *int64
		if row.FailoverEndTime > emptyFailoverEndTime {
			domainFailoverEndTime = common.Int64Ptr(row.FailoverEndTime)
		}
		domains = append(domains, &p.InternalGetDomainResponse{
			Info:                        row.Info,
			Config:                      row.Config,
			ReplicationConfig:           row.ReplicationConfig,
			IsGlobalDomain:              row.IsGlobalDomain,
			ConfigVersion:               row.ConfigVersion,
			FailoverVersion:             row.FailoverVersion,
			FailoverNotificationVersion: row.FailoverNotificationVersion,
			PreviousFailoverVersion:     row.PreviousFailoverVersion,
			FailoverEndTime:             domainFailoverEndTime,
			NotificationVersion:         row.NotificationVersion,
		})
	}

	return &p.InternalListDomainsResponse{
		Domains:       domains,
		NextPageToken: nextPageToken,
	}, nil
}

func (m *nosqlDomainManager) DeleteDomain(
	ctx context.Context,
	request *p.DeleteDomainRequest,
) error {
	return m.db.DeleteDomain(ctx, &request.ID, nil)
}

func (m *nosqlDomainManager) DeleteDomainByName(
	ctx context.Context,
	request *p.DeleteDomainByNameRequest,
) error {
	return m.db.DeleteDomain(ctx, nil, &request.Name)
}

func (m *nosqlDomainManager) GetMetadata(
	ctx context.Context,
) (*p.GetMetadataResponse, error) {
	notificationVersion, err := m.db.SelectDomainMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return &p.GetMetadataResponse{NotificationVersion: notificationVersion}, nil
}

func (m *nosqlDomainManager) updateMetadataBatch(batch *gocql.Batch, notificationVersion int64) {
	var nextVersion int64 = 1
	var currentVersion *int64
	if notificationVersion > 0 {
		nextVersion = notificationVersion + 1
		currentVersion = &notificationVersion
	}
	batch.Query(templateUpdateMetadataQueryWithinBatchV2,
		nextVersion,
		constDomainPartition,
		domainMetadataRecordName,
		currentVersion,
	)
}

// Copyright (c) 2020 Uber Technologies, Inc.
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
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
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
		`domains_partition, name, domain, config, replication_config, is_global_domain, config_version, failover_version, failover_notification_version, previous_failover_version, failover_end_time, last_updated_time, notification_version) ` +
		`VALUES(?, ?, ` + templateDomainInfoType + `, ` + templateDomainConfigType + `, ` + templateDomainReplicationConfigType + `, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS`

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
		`last_updated_time, ` +
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
		`last_updated_time = ?,` +
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
		`last_updated_time, ` +
		`notification_version ` +
		`FROM domains_by_name_v2 ` +
		`WHERE domains_partition = ? `
)

// Insert a new record to domain, return error if failed or already exists
// Must return conditionFailed error if domainName already exists
func (db *cdb) InsertDomain(
	ctx context.Context,
	row *nosqlplugin.DomainRow,
) error {
	query := db.session.Query(templateCreateDomainQuery, row.Info.ID, row.Info.Name).WithContext(ctx)
	applied, err := query.MapScanCAS(make(map[string]interface{}))
	if err != nil {
		return err
	}
	if !applied {
		return fmt.Errorf("CreateDomain operation failed because of uuid collision")
	}

	metadataNotificationVersion, err := db.SelectDomainMetadata(ctx)
	if err != nil {
		return err
	}

	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	failoverEndTime := emptyFailoverEndTime
	if row.FailoverEndTime != nil {
		failoverEndTime = row.FailoverEndTime.UnixNano()
	}
	batch.Query(templateCreateDomainByNameQueryWithinBatchV2,
		constDomainPartition,
		row.Info.Name,
		row.Info.ID,
		row.Info.Name,
		row.Info.Status,
		row.Info.Description,
		row.Info.OwnerEmail,
		row.Info.Data,
		common.DurationToDays(row.Config.Retention),
		row.Config.EmitMetric,
		row.Config.ArchivalBucket,
		row.Config.ArchivalStatus,
		row.Config.HistoryArchivalStatus,
		row.Config.HistoryArchivalURI,
		row.Config.VisibilityArchivalStatus,
		row.Config.VisibilityArchivalURI,
		row.Config.BadBinaries.Data,
		string(row.Config.BadBinaries.Encoding),
		row.ReplicationConfig.ActiveClusterName,
		p.SerializeClusterConfigs(row.ReplicationConfig.Clusters),
		row.IsGlobalDomain,
		row.ConfigVersion,
		row.FailoverVersion,
		p.InitialFailoverNotificationVersion,
		common.InitialPreviousFailoverVersion,
		failoverEndTime,
		row.LastUpdatedTime.UnixNano(),
		metadataNotificationVersion,
	)
	db.updateMetadataBatch(batch, metadataNotificationVersion)

	previous := make(map[string]interface{})
	applied, iter, err := db.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			_ = iter.Close()
		}
	}()

	if err != nil {
		return err
	}

	if !applied {
		// Domain already exist.  Delete orphan domain record before returning back to user
		if errDelete := db.session.Query(templateDeleteDomainQuery, row.Info.ID).WithContext(ctx).Exec(); errDelete != nil {
			db.logger.Warn("Unable to delete orphan domain record. Error", tag.Error(errDelete))
		}

		for {
			// first iter MapScan is done inside MapExecuteBatchCAS
			if domain, ok := previous["name"].(string); ok && domain == row.Info.Name {
				db.logger.Warn("Domain already exists", tag.WorkflowDomainName(domain))
				return &types.DomainAlreadyExistsError{
					Message: fmt.Sprintf("Domain %v already exists", previous["domain"]),
				}
			}

			previous = make(map[string]interface{})
			if !iter.MapScan(previous) {
				break
			}
		}

		db.logger.Warn("Create domain operation failed because of condition update failure on domain metadata record")
		return errConditionFailed
	}

	return nil
}

func (db *cdb) updateMetadataBatch(
	batch gocql.Batch,
	notificationVersion int64,
) {
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

// Update domain
func (db *cdb) UpdateDomain(
	ctx context.Context,
	row *nosqlplugin.DomainRow,
) error {
	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	failoverEndTime := emptyFailoverEndTime
	if row.FailoverEndTime != nil {
		failoverEndTime = row.FailoverEndTime.UnixNano()
	}
	batch.Query(templateUpdateDomainByNameQueryWithinBatchV2,
		row.Info.ID,
		row.Info.Name,
		row.Info.Status,
		row.Info.Description,
		row.Info.OwnerEmail,
		row.Info.Data,
		common.DurationToDays(row.Config.Retention),
		row.Config.EmitMetric,
		row.Config.ArchivalBucket,
		row.Config.ArchivalStatus,
		row.Config.HistoryArchivalStatus,
		row.Config.HistoryArchivalURI,
		row.Config.VisibilityArchivalStatus,
		row.Config.VisibilityArchivalURI,
		row.Config.BadBinaries.Data,
		string(row.Config.BadBinaries.Encoding),
		row.ReplicationConfig.ActiveClusterName,
		p.SerializeClusterConfigs(row.ReplicationConfig.Clusters),
		row.ConfigVersion,
		row.FailoverVersion,
		row.FailoverNotificationVersion,
		row.PreviousFailoverVersion,
		failoverEndTime,
		row.LastUpdatedTime.UnixNano(),
		row.NotificationVersion,
		constDomainPartition,
		row.Info.Name,
	)
	db.updateMetadataBatch(batch, row.NotificationVersion)

	previous := make(map[string]interface{})
	applied, iter, err := db.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return err
	}
	if !applied {
		return errConditionFailed
	}
	return nil
}

// Get one domain data, either by domainID or domainName
func (db *cdb) SelectDomain(
	ctx context.Context,
	domainID *string,
	domainName *string,
) (*nosqlplugin.DomainRow, error) {
	if domainID != nil && domainName != nil {
		return nil, fmt.Errorf("GetDomain operation failed.  Both ID and Name specified in request")
	} else if domainID == nil && domainName == nil {
		return nil, fmt.Errorf("GetDomain operation failed.  Both ID and Name are empty")
	}

	var query gocql.Query
	var err error
	if domainID != nil {
		query = db.session.Query(templateGetDomainQuery, domainID).WithContext(ctx)
		err = query.Scan(&domainName)
		if err != nil {
			return nil, err
		}
	}

	info := &p.DomainInfo{}
	config := &nosqlplugin.NoSQLInternalDomainConfig{}
	replicationConfig := &p.DomainReplicationConfig{}

	// because of encoding/types, we can't directly read from config struct
	var badBinariesData []byte
	var badBinariesDataEncoding string
	var replicationClusters []map[string]interface{}

	var failoverNotificationVersion int64
	var notificationVersion int64
	var failoverVersion int64
	var previousFailoverVersion int64
	var failoverEndTime int64
	var lastUpdatedTime int64
	var configVersion int64
	var isGlobalDomain bool
	var retentionDays int32

	query = db.session.Query(templateGetDomainByNameQueryV2, constDomainPartition, domainName).WithContext(ctx)
	err = query.Scan(
		&info.ID,
		&info.Name,
		&info.Status,
		&info.Description,
		&info.OwnerEmail,
		&info.Data,
		&retentionDays,
		&config.EmitMetric,
		&config.ArchivalBucket,
		&config.ArchivalStatus,
		&config.HistoryArchivalStatus,
		&config.HistoryArchivalURI,
		&config.VisibilityArchivalStatus,
		&config.VisibilityArchivalURI,
		&badBinariesData,
		&badBinariesDataEncoding,
		&replicationConfig.ActiveClusterName,
		&replicationClusters,
		&isGlobalDomain,
		&configVersion,
		&failoverVersion,
		&failoverNotificationVersion,
		&previousFailoverVersion,
		&failoverEndTime,
		&lastUpdatedTime,
		&notificationVersion,
	)

	if err != nil {
		return nil, err
	}

	config.BadBinaries = p.NewDataBlob(badBinariesData, common.EncodingType(badBinariesDataEncoding))
	config.Retention = common.DaysToDuration(retentionDays)
	replicationConfig.Clusters = p.DeserializeClusterConfigs(replicationClusters)

	dr := &nosqlplugin.DomainRow{
		Info:                        info,
		Config:                      config,
		ReplicationConfig:           replicationConfig,
		ConfigVersion:               configVersion,
		FailoverVersion:             failoverVersion,
		FailoverNotificationVersion: failoverNotificationVersion,
		PreviousFailoverVersion:     previousFailoverVersion,
		NotificationVersion:         notificationVersion,
		LastUpdatedTime:             time.Unix(0, lastUpdatedTime),
		IsGlobalDomain:              isGlobalDomain,
	}
	if failoverEndTime > emptyFailoverEndTime {
		dr.FailoverEndTime = common.TimePtr(time.Unix(0, failoverEndTime))
	}

	return dr, nil
}

// Get all domain data
func (db *cdb) SelectAllDomains(
	ctx context.Context,
	pageSize int,
	pageToken []byte,
) ([]*nosqlplugin.DomainRow, []byte, error) {
	var query gocql.Query
	query = db.session.Query(templateListDomainQueryV2, constDomainPartition).WithContext(ctx)
	iter := query.PageSize(pageSize).PageState(pageToken).Iter()
	if iter == nil {
		return nil, nil, &types.InternalServiceError{
			Message: "SelectAllDomains operation failed.  Not able to create query iterator.",
		}
	}

	var name string
	domain := &nosqlplugin.DomainRow{
		Info:              &p.DomainInfo{},
		Config:            &nosqlplugin.NoSQLInternalDomainConfig{},
		ReplicationConfig: &p.DomainReplicationConfig{},
	}
	var replicationClusters []map[string]interface{}
	var badBinariesData []byte
	var badBinariesDataEncoding string
	var retentionDays int32
	var failoverEndTime int64
	var lastUpdateTime int64
	var rows []*nosqlplugin.DomainRow
	for iter.Scan(
		&name,
		&domain.Info.ID,
		&domain.Info.Name,
		&domain.Info.Status,
		&domain.Info.Description,
		&domain.Info.OwnerEmail,
		&domain.Info.Data,
		&retentionDays,
		&domain.Config.EmitMetric,
		&domain.Config.ArchivalBucket,
		&domain.Config.ArchivalStatus,
		&domain.Config.HistoryArchivalStatus,
		&domain.Config.HistoryArchivalURI,
		&domain.Config.VisibilityArchivalStatus,
		&domain.Config.VisibilityArchivalURI,
		&badBinariesData,
		&badBinariesDataEncoding,
		&domain.ReplicationConfig.ActiveClusterName,
		&replicationClusters,
		&domain.IsGlobalDomain,
		&domain.ConfigVersion,
		&domain.FailoverVersion,
		&domain.FailoverNotificationVersion,
		&domain.PreviousFailoverVersion,
		&failoverEndTime,
		&lastUpdateTime,
		&domain.NotificationVersion,
	) {
		if name != domainMetadataRecordName {
			// do not include the metadata record
			domain.Config.BadBinaries = p.NewDataBlob(badBinariesData, common.EncodingType(badBinariesDataEncoding))
			domain.ReplicationConfig.Clusters = p.DeserializeClusterConfigs(replicationClusters)
			domain.Config.Retention = common.DaysToDuration(retentionDays)
			domain.LastUpdatedTime = time.Unix(0, lastUpdateTime)
			if failoverEndTime > emptyFailoverEndTime {
				domain.FailoverEndTime = common.TimePtr(time.Unix(0, failoverEndTime))
			}
			rows = append(rows, domain)
		}
		replicationClusters = []map[string]interface{}{}
		badBinariesData = []byte("")
		badBinariesDataEncoding = ""
		failoverEndTime = 0
		lastUpdateTime = 0
		retentionDays = 0
		domain = &nosqlplugin.DomainRow{
			Info:              &p.DomainInfo{},
			Config:            &nosqlplugin.NoSQLInternalDomainConfig{},
			ReplicationConfig: &p.DomainReplicationConfig{},
		}
	}

	nextPageToken := iter.PageState()
	if err := iter.Close(); err != nil {
		return nil, nil, err
	}
	return rows, nextPageToken, nil
}

//  Delete a domain, either by domainID or domainName
func (db *cdb) DeleteDomain(
	ctx context.Context,
	domainID *string,
	domainName *string,
) error {
	if domainName == nil && domainID == nil {
		return fmt.Errorf("must provide either domainID or domainName")
	}

	if domainName == nil {
		query := db.session.Query(templateGetDomainQuery, domainID).WithContext(ctx)
		var name string
		err := query.Scan(&name)
		if err != nil {
			if db.client.IsNotFoundError(err) {
				return nil
			}
			return err
		}
		domainName = common.StringPtr(name)
	} else {
		var ID string
		query := db.session.Query(templateGetDomainByNameQueryV2, constDomainPartition, *domainName).WithContext(ctx)
		err := query.Scan(&ID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			if db.client.IsNotFoundError(err) {
				return nil
			}
			return err
		}
		domainID = common.StringPtr(ID)
	}

	return db.deleteDomain(ctx, *domainName, *domainID)
}

func (db *cdb) SelectDomainMetadata(
	ctx context.Context,
) (int64, error) {
	var notificationVersion int64
	query := db.session.Query(templateGetMetadataQueryV2, constDomainPartition, domainMetadataRecordName)
	err := query.Scan(&notificationVersion)
	if err != nil {
		if db.client.IsNotFoundError(err) {
			// this error can be thrown in the very beginning,
			// i.e. when domains_by_name_v2 is initialized
			// TODO ??? really????
			return 0, nil
		}
		return -1, err
	}
	return notificationVersion, nil
}

func (db *cdb) deleteDomain(
	ctx context.Context,
	name, ID string,
) error {
	query := db.session.Query(templateDeleteDomainByNameQueryV2, constDomainPartition, name).WithContext(ctx)
	if err := query.Exec(); err != nil {
		return err
	}

	query = db.session.Query(templateDeleteDomainQuery, ID).WithContext(ctx)
	return query.Exec()
}

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

package domain

import (
	"context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

// NOTE: the counterpart of domain replication receiving logic is in service/worker package

type (
	// Replicator is the interface which can replicate the domain
	Replicator interface {
		HandleTransmissionTask(
			ctx context.Context,
			domainOperation types.DomainOperation,
			info *persistence.DomainInfo,
			config *persistence.DomainConfig,
			replicationConfig *persistence.DomainReplicationConfig,
			configVersion int64,
			failoverVersion int64,
			previousFailoverVersion int64,
			isGlobalDomainEnabled bool,
		) error
	}

	domainReplicatorImpl struct {
		replicationMessageSink messaging.Producer
		logger                 log.Logger
	}
)

// NewDomainReplicator create a new instance of domain replicator
func NewDomainReplicator(replicationMessageSink messaging.Producer, logger log.Logger) Replicator {
	return &domainReplicatorImpl{
		replicationMessageSink: replicationMessageSink,
		logger:                 logger,
	}
}

// HandleTransmissionTask handle transmission of the domain replication task
func (domainReplicator *domainReplicatorImpl) HandleTransmissionTask(
	ctx context.Context,
	domainOperation types.DomainOperation,
	info *persistence.DomainInfo,
	config *persistence.DomainConfig,
	replicationConfig *persistence.DomainReplicationConfig,
	configVersion int64,
	failoverVersion int64,
	previousFailoverVersion int64,
	isGlobalDomainEnabled bool,
) error {

	if !isGlobalDomainEnabled {
		domainReplicator.logger.Warn("Should not replicate non global domain", tag.WorkflowDomainID(info.ID))
		return nil
	}

	status, err := domainReplicator.convertDomainStatusToThrift(info.Status)
	if err != nil {
		return err
	}

	taskType := types.ReplicationTaskTypeDomain
	task := &types.DomainTaskAttributes{
		DomainOperation: &domainOperation,
		ID:              common.StringPtr(info.ID),
		Info: &types.DomainInfo{
			Name:        common.StringPtr(info.Name),
			Status:      status,
			Description: common.StringPtr(info.Description),
			OwnerEmail:  common.StringPtr(info.OwnerEmail),
			Data:        info.Data,
		},
		Config: &types.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(config.Retention),
			EmitMetric:                             common.BoolPtr(config.EmitMetric),
			HistoryArchivalStatus:                  config.HistoryArchivalStatus.Ptr(),
			HistoryArchivalURI:                     common.StringPtr(config.HistoryArchivalURI),
			VisibilityArchivalStatus:               config.VisibilityArchivalStatus.Ptr(),
			VisibilityArchivalURI:                  common.StringPtr(config.VisibilityArchivalURI),
			BadBinaries:                            &config.BadBinaries,
		},
		ReplicationConfig: &types.DomainReplicationConfiguration{
			ActiveClusterName: common.StringPtr(replicationConfig.ActiveClusterName),
			Clusters:          domainReplicator.convertClusterReplicationConfigToThrift(replicationConfig.Clusters),
		},
		ConfigVersion:           common.Int64Ptr(configVersion),
		FailoverVersion:         common.Int64Ptr(failoverVersion),
		PreviousFailoverVersion: common.Int64Ptr(previousFailoverVersion),
	}

	return domainReplicator.replicationMessageSink.Publish(
		ctx,
		&types.ReplicationTask{
			TaskType:             &taskType,
			DomainTaskAttributes: task,
		})
}

func (domainReplicator *domainReplicatorImpl) convertClusterReplicationConfigToThrift(
	input []*persistence.ClusterReplicationConfig,
) []*types.ClusterReplicationConfiguration {
	output := []*types.ClusterReplicationConfiguration{}
	for _, cluster := range input {
		clusterName := common.StringPtr(cluster.ClusterName)
		output = append(output, &types.ClusterReplicationConfiguration{ClusterName: clusterName})
	}
	return output
}

func (domainReplicator *domainReplicatorImpl) convertDomainStatusToThrift(input int) (*types.DomainStatus, error) {
	switch input {
	case persistence.DomainStatusRegistered:
		output := types.DomainStatusRegistered
		return &output, nil
	case persistence.DomainStatusDeprecated:
		output := types.DomainStatusDeprecated
		return &output, nil
	default:
		return nil, ErrInvalidDomainStatus
	}
}

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
	"context"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
)

type (

	// metadataManagerImpl implements MetadataManager based on MetadataStore and PayloadSerializer
	metadataManagerImpl struct {
		serializer  PayloadSerializer
		persistence MetadataStore
		logger      log.Logger
	}
)

var _ MetadataManager = (*metadataManagerImpl)(nil)

//NewMetadataManagerImpl returns new MetadataManager
func NewMetadataManagerImpl(persistence MetadataStore, logger log.Logger) MetadataManager {
	return &metadataManagerImpl{
		serializer:  NewPayloadSerializer(),
		persistence: persistence,
		logger:      logger,
	}
}

func (m *metadataManagerImpl) GetName() string {
	return m.persistence.GetName()
}

func (m *metadataManagerImpl) CreateDomain(
	ctx context.Context,
	request *CreateDomainRequest,
) (*CreateDomainResponse, error) {
	dc, err := m.serializeDomainConfig(request.Config)
	if err != nil {
		return nil, err
	}

	internalRequest := &InternalCreateDomainRequest{
		Info:              m.toInternalDomainInfo(request.Info),
		Config:            &dc,
		ReplicationConfig: m.toInternalDomainReplicationConfig(request.ReplicationConfig),
		IsGlobalDomain:    request.IsGlobalDomain,
		ConfigVersion:     request.ConfigVersion,
		FailoverVersion:   request.FailoverVersion,
	}

	resp, err := m.persistence.CreateDomain(ctx, internalRequest)
	if err != nil {
		return nil, err
	}

	return &CreateDomainResponse{
		ID: resp.ID,
	}, nil
}

func (m *metadataManagerImpl) GetDomain(
	ctx context.Context,
	request *GetDomainRequest,
) (*GetDomainResponse, error) {
	internalRequest := &InternalGetDomainRequest{
		ID:   request.ID,
		Name: request.Name,
	}
	resp, err := m.persistence.GetDomain(ctx, internalRequest)
	if err != nil {
		return nil, err
	}

	dc, err := m.deserializeDomainConfig(resp.Config)
	if err != nil {
		return nil, err
	}

	return &GetDomainResponse{
		Info:                        m.fromInternalDomainInfo(resp.Info),
		Config:                      &dc,
		ReplicationConfig:           m.fromInternalDomainReplicationConfig(resp.ReplicationConfig),
		IsGlobalDomain:              resp.IsGlobalDomain,
		ConfigVersion:               resp.ConfigVersion,
		FailoverVersion:             resp.FailoverVersion,
		FailoverNotificationVersion: resp.FailoverNotificationVersion,
		PreviousFailoverVersion:     resp.PreviousFailoverVersion,
		FailoverEndTime:             resp.FailoverEndTime,
		NotificationVersion:         resp.NotificationVersion,
	}, nil
}

func (m *metadataManagerImpl) UpdateDomain(
	ctx context.Context,
	request *UpdateDomainRequest,
) error {
	dc, err := m.serializeDomainConfig(request.Config)
	if err != nil {
		return err
	}
	return m.persistence.UpdateDomain(ctx, &InternalUpdateDomainRequest{
		Info:                        m.toInternalDomainInfo(request.Info),
		Config:                      &dc,
		ReplicationConfig:           m.toInternalDomainReplicationConfig(request.ReplicationConfig),
		ConfigVersion:               request.ConfigVersion,
		FailoverVersion:             request.FailoverVersion,
		FailoverNotificationVersion: request.FailoverNotificationVersion,
		PreviousFailoverVersion:     request.PreviousFailoverVersion,
		FailoverEndTime:             request.FailoverEndTime,
		NotificationVersion:         request.NotificationVersion,
	})
}

func (m *metadataManagerImpl) DeleteDomain(
	ctx context.Context,
	request *DeleteDomainRequest,
) error {
	return m.persistence.DeleteDomain(ctx, &InternalDeleteDomainRequest{
		ID: request.ID,
	})
}

func (m *metadataManagerImpl) DeleteDomainByName(
	ctx context.Context,
	request *DeleteDomainByNameRequest,
) error {
	return m.persistence.DeleteDomainByName(ctx, &InternalDeleteDomainByNameRequest{
		Name: request.Name,
	})
}

func (m *metadataManagerImpl) ListDomains(
	ctx context.Context,
	request *ListDomainsRequest,
) (*ListDomainsResponse, error) {
	internalRequest := &InternalListDomainRequest{
		PageSize:      request.PageSize,
		NextPageToken: request.NextPageToken,
	}
	resp, err := m.persistence.ListDomains(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	domains := make([]*GetDomainResponse, 0, len(resp.Domains))
	for _, d := range resp.Domains {
		dc, err := m.deserializeDomainConfig(d.Config)
		if err != nil {
			return nil, err
		}
		domains = append(domains, &GetDomainResponse{
			Info:                        m.fromInternalDomainInfo(d.Info),
			Config:                      &dc,
			ReplicationConfig:           m.fromInternalDomainReplicationConfig(d.ReplicationConfig),
			IsGlobalDomain:              d.IsGlobalDomain,
			ConfigVersion:               d.ConfigVersion,
			FailoverVersion:             d.FailoverVersion,
			FailoverNotificationVersion: d.FailoverNotificationVersion,
			FailoverEndTime:             d.FailoverEndTime,
			PreviousFailoverVersion:     d.PreviousFailoverVersion,
			NotificationVersion:         d.NotificationVersion,
		})
	}
	return &ListDomainsResponse{
		Domains:       domains,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (m *metadataManagerImpl) toInternalDomainInfo(domainInfo *DomainInfo) *InternalDomainInfo {
	if domainInfo == nil {
		return nil
	}
	return &InternalDomainInfo{
		ID:          domainInfo.ID,
		Name:        domainInfo.Name,
		Status:      domainInfo.Status,
		Description: domainInfo.Description,
		OwnerEmail:  domainInfo.OwnerEmail,
		Data:        domainInfo.Data,
	}
}

func (m *metadataManagerImpl) fromInternalDomainInfo(internalDomainInfo *InternalDomainInfo) *DomainInfo {
	if internalDomainInfo == nil {
		return nil
	}
	return &DomainInfo{
		ID:          internalDomainInfo.ID,
		Name:        internalDomainInfo.Name,
		Status:      internalDomainInfo.Status,
		Description: internalDomainInfo.Description,
		OwnerEmail:  internalDomainInfo.OwnerEmail,
		Data:        internalDomainInfo.Data,
	}
}

func (m *metadataManagerImpl) toInternalDomainReplicationConfig(domainReplicationConfig *DomainReplicationConfig) *InternalDomainReplicationConfig {
	if domainReplicationConfig == nil {
		return nil
	}
	var internalClusters []*InternalClusterReplicationConfig
	for _, cluster := range domainReplicationConfig.Clusters {
		internalClusters = append(internalClusters, &InternalClusterReplicationConfig{
			ClusterName: cluster.ClusterName,
		})
	}
	return &InternalDomainReplicationConfig{
		ActiveClusterName: domainReplicationConfig.ActiveClusterName,
		Clusters:          internalClusters,
	}
}

func (m *metadataManagerImpl) fromInternalDomainReplicationConfig(internalDomainReplicationConfig *InternalDomainReplicationConfig) *DomainReplicationConfig {
	if internalDomainReplicationConfig == nil {
		return nil
	}
	var internalClusters []*ClusterReplicationConfig
	for _, cluster := range internalDomainReplicationConfig.Clusters {
		internalClusters = append(internalClusters, &ClusterReplicationConfig{
			ClusterName: cluster.ClusterName,
		})
	}
	return &DomainReplicationConfig{
		ActiveClusterName: internalDomainReplicationConfig.ActiveClusterName,
		Clusters:          internalClusters,
	}
}

func (m *metadataManagerImpl) serializeDomainConfig(c *DomainConfig) (InternalDomainConfig, error) {
	if c == nil {
		return InternalDomainConfig{}, nil
	}
	if c.BadBinaries.Binaries == nil {
		c.BadBinaries.Binaries = map[string]*shared.BadBinaryInfo{}
	}
	badBinaries, err := m.serializer.SerializeBadBinaries(&c.BadBinaries, common.EncodingTypeThriftRW)
	if err != nil {
		return InternalDomainConfig{}, err
	}
	return InternalDomainConfig{
		Retention:                c.Retention,
		EmitMetric:               c.EmitMetric,
		HistoryArchivalStatus:    c.HistoryArchivalStatus,
		HistoryArchivalURI:       c.HistoryArchivalURI,
		VisibilityArchivalStatus: c.VisibilityArchivalStatus,
		VisibilityArchivalURI:    c.VisibilityArchivalURI,
		BadBinaries:              badBinaries,
	}, nil
}

func (m *metadataManagerImpl) deserializeDomainConfig(ic *InternalDomainConfig) (DomainConfig, error) {
	if ic == nil {
		return DomainConfig{}, nil
	}
	badBinaries, err := m.serializer.DeserializeBadBinaries(ic.BadBinaries)
	if err != nil {
		return DomainConfig{}, err
	}
	if badBinaries.Binaries == nil {
		badBinaries.Binaries = map[string]*shared.BadBinaryInfo{}
	}
	return DomainConfig{
		Retention:                ic.Retention,
		EmitMetric:               ic.EmitMetric,
		HistoryArchivalStatus:    ic.HistoryArchivalStatus,
		HistoryArchivalURI:       ic.HistoryArchivalURI,
		VisibilityArchivalStatus: ic.VisibilityArchivalStatus,
		VisibilityArchivalURI:    ic.VisibilityArchivalURI,
		BadBinaries:              *badBinaries,
	}, nil
}

func (m *metadataManagerImpl) GetMetadata(
	ctx context.Context,
) (*GetMetadataResponse, error) {
	resp, err := m.persistence.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return &GetMetadataResponse{
		NotificationVersion: resp.NotificationVersion,
	}, nil
}

func (m *metadataManagerImpl) Close() {
	m.persistence.Close()
}

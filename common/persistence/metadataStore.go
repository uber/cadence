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

	"github.com/uber/cadence/common"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types/mapper/thrift"
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
	dc, err := m.toInternalDomainConfig(request.Config)
	if err != nil {
		return nil, err
	}
	return m.persistence.CreateDomain(ctx, &InternalCreateDomainRequest{
		Info:              request.Info,
		Config:            &dc,
		ReplicationConfig: request.ReplicationConfig,
		IsGlobalDomain:    request.IsGlobalDomain,
		ConfigVersion:     request.ConfigVersion,
		FailoverVersion:   request.FailoverVersion,
		LastUpdatedTime:   request.LastUpdatedTime,
	})
}

func (m *metadataManagerImpl) GetDomain(
	ctx context.Context,
	request *GetDomainRequest,
) (*GetDomainResponse, error) {
	resp, err := m.persistence.GetDomain(ctx, request)
	if err != nil {
		return nil, err
	}

	dc, err := m.fromInternalDomainConfig(resp.Config)
	if err != nil {
		return nil, err
	}

	return &GetDomainResponse{
		Info:                        resp.Info,
		Config:                      &dc,
		ReplicationConfig:           resp.ReplicationConfig,
		IsGlobalDomain:              resp.IsGlobalDomain,
		ConfigVersion:               resp.ConfigVersion,
		FailoverVersion:             resp.FailoverVersion,
		FailoverNotificationVersion: resp.FailoverNotificationVersion,
		PreviousFailoverVersion:     resp.PreviousFailoverVersion,
		FailoverEndTime:             resp.FailoverEndTime,
		LastUpdatedTime:             resp.LastUpdatedTime,
		NotificationVersion:         resp.NotificationVersion,
	}, nil
}

func (m *metadataManagerImpl) UpdateDomain(
	ctx context.Context,
	request *UpdateDomainRequest,
) error {
	dc, err := m.toInternalDomainConfig(request.Config)
	if err != nil {
		return err
	}
	return m.persistence.UpdateDomain(ctx, &InternalUpdateDomainRequest{
		Info:                        request.Info,
		Config:                      &dc,
		ReplicationConfig:           request.ReplicationConfig,
		ConfigVersion:               request.ConfigVersion,
		FailoverVersion:             request.FailoverVersion,
		FailoverNotificationVersion: request.FailoverNotificationVersion,
		PreviousFailoverVersion:     request.PreviousFailoverVersion,
		FailoverEndTime:             request.FailoverEndTime,
		LastUpdatedTime:             request.LastUpdatedTime,
		NotificationVersion:         request.NotificationVersion,
	})
}

func (m *metadataManagerImpl) DeleteDomain(
	ctx context.Context,
	request *DeleteDomainRequest,
) error {
	return m.persistence.DeleteDomain(ctx, request)
}

func (m *metadataManagerImpl) DeleteDomainByName(
	ctx context.Context,
	request *DeleteDomainByNameRequest,
) error {
	return m.persistence.DeleteDomainByName(ctx, request)
}

func (m *metadataManagerImpl) ListDomains(
	ctx context.Context,
	request *ListDomainsRequest,
) (*ListDomainsResponse, error) {
	resp, err := m.persistence.ListDomains(ctx, request)
	if err != nil {
		return nil, err
	}
	domains := make([]*GetDomainResponse, 0, len(resp.Domains))
	for _, d := range resp.Domains {
		dc, err := m.fromInternalDomainConfig(d.Config)
		if err != nil {
			return nil, err
		}
		domains = append(domains, &GetDomainResponse{
			Info:                        d.Info,
			Config:                      &dc,
			ReplicationConfig:           d.ReplicationConfig,
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

func (m *metadataManagerImpl) toInternalDomainConfig(c *DomainConfig) (InternalDomainConfig, error) {
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
		HistoryArchivalStatus:    *thrift.ToArchivalStatus(&c.HistoryArchivalStatus),
		HistoryArchivalURI:       c.HistoryArchivalURI,
		VisibilityArchivalStatus: *thrift.ToArchivalStatus(&c.VisibilityArchivalStatus),
		VisibilityArchivalURI:    c.VisibilityArchivalURI,
		BadBinaries:              badBinaries,
	}, nil
}

func (m *metadataManagerImpl) fromInternalDomainConfig(ic *InternalDomainConfig) (DomainConfig, error) {
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
		HistoryArchivalStatus:    *thrift.FromArchivalStatus(&ic.HistoryArchivalStatus),
		HistoryArchivalURI:       ic.HistoryArchivalURI,
		VisibilityArchivalStatus: *thrift.FromArchivalStatus(&ic.VisibilityArchivalStatus),
		VisibilityArchivalURI:    ic.VisibilityArchivalURI,
		BadBinaries:              *badBinaries,
	}, nil
}

func (m *metadataManagerImpl) GetMetadata(
	ctx context.Context,
) (*GetMetadataResponse, error) {
	return m.persistence.GetMetadata(ctx)
}

func (m *metadataManagerImpl) Close() {
	m.persistence.Close()
}

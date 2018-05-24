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

package cache

import (
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"
)

type (
	// TODO, we should migrate the non global domain to new table, see #773
	metadataManagerProxy struct {
		metadataMgr   persistence.MetadataManager
		metadataMgrV2 persistence.MetadataManager
	}
)

// NewMetadataManagerProxy is used for merging the functionality the v1 and v2 MetadataManager
func NewMetadataManagerProxy(metadataMgr persistence.MetadataManager, metadataMgrV2 persistence.MetadataManager) persistence.MetadataManager {
	return &metadataManagerProxy{
		metadataMgr:   metadataMgr,
		metadataMgrV2: metadataMgrV2,
	}
}

func (m *metadataManagerProxy) GetDomain(request *persistence.GetDomainRequest) (*persistence.GetDomainResponse, error) {
	resp, err := m.metadataMgrV2.GetDomain(request)
	if err == nil {
		return resp, nil
	}

	// if entity not exist, try the old table
	if _, ok := err.(*workflow.EntityNotExistsError); !ok {
		return nil, err
	}

	resp, err = m.metadataMgr.GetDomain(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *metadataManagerProxy) ListDomain(request *persistence.ListDomainRequest) (*persistence.ListDomainResponse, error) {
	return m.metadataMgrV2.ListDomain(request)
}

func (m *metadataManagerProxy) GetMetadata() (int64, error) {
	return m.metadataMgrV2.GetMetadata()
}

func (m *metadataManagerProxy) Close() {

}

func (m *metadataManagerProxy) CreateDomain(request *persistence.CreateDomainRequest) (*persistence.CreateDomainResponse, error) {
	panic("metadataManagerProxy do not support create domain operation.")
}

func (m *metadataManagerProxy) UpdateDomain(request *persistence.UpdateDomainRequest) error {
	panic("metadataManagerProxy do not support update domain operation.")
}

func (m *metadataManagerProxy) DeleteDomain(request *persistence.DeleteDomainRequest) error {
	panic("metadataManagerProxy do not support delete domain operation.")
}

func (m *metadataManagerProxy) DeleteDomainByName(request *persistence.DeleteDomainByNameRequest) error {
	panic("metadataManagerProxy do not support delete domain by name operation.")
}

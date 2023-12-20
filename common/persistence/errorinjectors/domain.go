// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package errorinjectors

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type domainClient struct {
	persistence persistence.DomainManager
	errorRate   float64
	logger      log.Logger
}

// NewDomainClient creates an error injection DomainManager client to manage metadata
func NewDomainClient(
	persistence persistence.DomainManager,
	errorRate float64,
	logger log.Logger,
) persistence.DomainManager {
	return &domainClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *domainClient) GetName() string {
	return p.persistence.GetName()
}

func (p *domainClient) CreateDomain(
	ctx context.Context,
	request *persistence.CreateDomainRequest,
) (*persistence.CreateDomainResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CreateDomainResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CreateDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCreateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *domainClient) GetDomain(
	ctx context.Context,
	request *persistence.GetDomainRequest,
) (*persistence.GetDomainResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetDomainResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *domainClient) UpdateDomain(
	ctx context.Context,
	request *persistence.UpdateDomainRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *domainClient) DeleteDomain(
	ctx context.Context,
	request *persistence.DeleteDomainRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *domainClient) DeleteDomainByName(
	ctx context.Context,
	request *persistence.DeleteDomainByNameRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteDomainByName(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteDomainByName,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *domainClient) ListDomains(
	ctx context.Context,
	request *persistence.ListDomainsRequest,
) (*persistence.ListDomainsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListDomainsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListDomains(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListDomains,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *domainClient) GetMetadata(
	ctx context.Context,
) (*persistence.GetMetadataResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetMetadataResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetMetadata(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetMetadata,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *domainClient) Close() {
	p.persistence.Close()
}

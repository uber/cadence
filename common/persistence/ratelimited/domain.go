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

package ratelimited

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

type domainClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.DomainManager
	logger      log.Logger
}

// NewDomainClient creates a DomainManager client to manage metadata
func NewDomainClient(
	persistence persistence.DomainManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.DomainManager {
	return &domainClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
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
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateDomain(ctx, request)
	return response, err
}

func (p *domainClient) GetDomain(
	ctx context.Context,
	request *persistence.GetDomainRequest,
) (*persistence.GetDomainResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetDomain(ctx, request)
	return response, err
}

func (p *domainClient) UpdateDomain(
	ctx context.Context,
	request *persistence.UpdateDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpdateDomain(ctx, request)
	return err
}

func (p *domainClient) DeleteDomain(
	ctx context.Context,
	request *persistence.DeleteDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomain(ctx, request)
	return err
}

func (p *domainClient) DeleteDomainByName(
	ctx context.Context,
	request *persistence.DeleteDomainByNameRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomainByName(ctx, request)
	return err
}

func (p *domainClient) ListDomains(
	ctx context.Context,
	request *persistence.ListDomainsRequest,
) (*persistence.ListDomainsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListDomains(ctx, request)
	return response, err
}

func (p *domainClient) GetMetadata(
	ctx context.Context,
) (*persistence.GetMetadataResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetMetadata(ctx)
	return response, err
}

func (p *domainClient) Close() {
	p.persistence.Close()
}

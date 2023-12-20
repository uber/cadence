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

type configStoreRateLimitedPersistenceClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.ConfigStoreManager
	logger      log.Logger
}

// NewConfigStorePersistenceRateLimitedClient creates a client to manage config store
func NewConfigStorePersistenceRateLimitedClient(
	persistence persistence.ConfigStoreManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.ConfigStoreManager {
	return &configStoreRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *configStoreRateLimitedPersistenceClient) FetchDynamicConfig(ctx context.Context, configType persistence.ConfigType) (*persistence.FetchDynamicConfigResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.FetchDynamicConfig(ctx, configType)
}

func (p *configStoreRateLimitedPersistenceClient) UpdateDynamicConfig(ctx context.Context, request *persistence.UpdateDynamicConfigRequest, configType persistence.ConfigType) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.UpdateDynamicConfig(ctx, request, configType)
}

func (p *configStoreRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

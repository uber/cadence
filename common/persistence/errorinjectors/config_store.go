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

type configStoreClient struct {
	persistence persistence.ConfigStoreManager
	errorRate   float64
	logger      log.Logger
}

// NewConfigStoreClient creates an error injection client to manage config store
func NewConfigStoreClient(
	persistence persistence.ConfigStoreManager,
	errorRate float64,
	logger log.Logger,
) persistence.ConfigStoreManager {
	return &configStoreClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *configStoreClient) FetchDynamicConfig(ctx context.Context, cfgType persistence.ConfigType) (*persistence.FetchDynamicConfigResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.FetchDynamicConfigResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.FetchDynamicConfig(ctx, cfgType)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationFetchDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *configStoreClient) UpdateDynamicConfig(ctx context.Context, request *persistence.UpdateDynamicConfigRequest, cfgType persistence.ConfigType) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateDynamicConfig(ctx, request, cfgType)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *configStoreClient) Close() {
	p.persistence.Close()
}

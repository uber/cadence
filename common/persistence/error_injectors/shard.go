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

package error_injectors

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type shardClient struct {
	persistence persistence.ShardManager
	errorRate   float64
	logger      log.Logger
}

// NewShardClient creates an error injection client to manage shards.
func NewShardClient(
	persistence persistence.ShardManager,
	errorRate float64,
	logger log.Logger,
) persistence.ShardManager {
	return &shardClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *shardClient) GetName() string {
	return p.persistence.GetName()
}

func (p *shardClient) CreateShard(
	ctx context.Context,
	request *persistence.CreateShardRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CreateShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationCreateShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return fakeErr
	}
	return persistenceErr
}

func (p *shardClient) GetShard(
	ctx context.Context,
	request *persistence.GetShardRequest,
) (*persistence.GetShardResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetShardResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationGetShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *shardClient) UpdateShard(
	ctx context.Context,
	request *persistence.UpdateShardRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationUpdateShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return fakeErr
	}
	return persistenceErr
}

func (p *shardClient) Close() {
	p.persistence.Close()
}

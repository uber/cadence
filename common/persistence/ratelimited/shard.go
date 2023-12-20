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

type shardClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.ShardManager
	logger      log.Logger
}

// NewShardClient creates a client to manage shards
func NewShardClient(
	persistence persistence.ShardManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.ShardManager {
	return &shardClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
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
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateShard(ctx, request)
	return err
}

func (p *shardClient) GetShard(
	ctx context.Context,
	request *persistence.GetShardRequest,
) (*persistence.GetShardResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetShard(ctx, request)
	return response, err
}

func (p *shardClient) UpdateShard(
	ctx context.Context,
	request *persistence.UpdateShardRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpdateShard(ctx, request)
	return err
}

func (p *shardClient) Close() {
	p.persistence.Close()
}

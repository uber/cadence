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

type historyClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.HistoryManager
	logger      log.Logger
}

// NewHistoryClient creates a HistoryManager client to manage workflow execution history
func NewHistoryClient(
	persistence persistence.HistoryManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.HistoryManager {
	return &historyClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *historyClient) GetName() string {
	return p.persistence.GetName()
}

func (p *historyClient) Close() {
	p.persistence.Close()
}

// AppendHistoryNodes add(or override) a node to a history branch
func (p *historyClient) AppendHistoryNodes(
	ctx context.Context,
	request *persistence.AppendHistoryNodesRequest,
) (*persistence.AppendHistoryNodesResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.AppendHistoryNodes(ctx, request)
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyClient) ReadHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranch(ctx, request)
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchByBatchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranchByBatch(ctx, request)
	return response, err
}

// ReadRawHistoryBranch returns history node data for a branch
func (p *historyClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadRawHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadRawHistoryBranch(ctx, request)
	return response, err
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyClient) ForkHistoryBranch(
	ctx context.Context,
	request *persistence.ForkHistoryBranchRequest,
) (*persistence.ForkHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ForkHistoryBranch(ctx, request)
	return response, err
}

// DeleteHistoryBranch removes a branch
func (p *historyClient) DeleteHistoryBranch(
	ctx context.Context,
	request *persistence.DeleteHistoryBranchRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	err := p.persistence.DeleteHistoryBranch(ctx, request)
	return err
}

// GetHistoryTree returns all branch information of a tree
func (p *historyClient) GetHistoryTree(
	ctx context.Context,
	request *persistence.GetHistoryTreeRequest,
) (*persistence.GetHistoryTreeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.GetHistoryTree(ctx, request)
	return response, err
}

func (p *historyClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *persistence.GetAllHistoryTreeBranchesRequest,
) (*persistence.GetAllHistoryTreeBranchesResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.GetAllHistoryTreeBranches(ctx, request)
	return response, err
}

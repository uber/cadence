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

type historyClient struct {
	persistence persistence.HistoryManager
	errorRate   float64
	logger      log.Logger
}

// NewHistoryClient creates an error injection HistoryManager client to manage workflow execution history
func NewHistoryClient(
	persistence persistence.HistoryManager,
	errorRate float64,
	logger log.Logger,
) persistence.HistoryManager {
	return &historyClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *historyClient) GetName() string {
	return p.persistence.GetName()
}

// AppendHistoryNodes add(or override) a node to a history branch
func (p *historyClient) AppendHistoryNodes(
	ctx context.Context,
	request *persistence.AppendHistoryNodesRequest,
) (*persistence.AppendHistoryNodesResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.AppendHistoryNodesResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.AppendHistoryNodes(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationAppendHistoryNodes,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyClient) ReadHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchByBatchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadHistoryBranchByBatchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadHistoryBranchByBatch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadHistoryBranchByBatch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadRawHistoryBranch returns history node data for a branch
func (p *historyClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadRawHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadRawHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadRawHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadRawHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyClient) ForkHistoryBranch(
	ctx context.Context,
	request *persistence.ForkHistoryBranchRequest,
) (*persistence.ForkHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ForkHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ForkHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationForkHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// DeleteHistoryBranch removes a branch
func (p *historyClient) DeleteHistoryBranch(
	ctx context.Context,
	request *persistence.DeleteHistoryBranchRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

// GetHistoryTree returns all branch information of a tree
func (p *historyClient) GetHistoryTree(
	ctx context.Context,
	request *persistence.GetHistoryTreeRequest,
) (*persistence.GetHistoryTreeResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetHistoryTreeResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetHistoryTree(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetHistoryTree,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *historyClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *persistence.GetAllHistoryTreeBranchesRequest,
) (*persistence.GetAllHistoryTreeBranchesResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetAllHistoryTreeBranchesResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetAllHistoryTreeBranches(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetAllHistoryTreeBranches,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *historyClient) Close() {
	p.persistence.Close()
}

// Copyright (c) 2019 Uber Technologies, Inc.
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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination nDCHistoryRereplicator_mock.go

package xdc

import (
	"context"
	"time"

	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	a "github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	// nDCHistoryReplicationFn provides the functionality to deliver replication raw history request to history
	// the provided func should be thread safe
	nDCHistoryReplicationFn func(ctx context.Context, request *history.ReplicateEventsV2Request) error

	// NDCHistoryRereplicator is the interface for resending history events to remote
	NDCHistoryRereplicator interface {
		// SendSingleWorkflowHistory sends multiple run IDs's history events to remote
		SendSingleWorkflowHistory(
			domainID string,
			workflowID string,
			runID string,
			startEventID *int64,
			startEventVersion *int64,
			endEventID *int64,
			endEventVersion *int64,
		) error
	}

	// NDCHistoryRereplicatorImpl is the implementation of NDCHistoryRereplicator
	NDCHistoryRereplicatorImpl struct {
		targetClusterName    string
		domainCache          cache.DomainCache
		adminClient          a.Client
		historyReplicationFn nDCHistoryReplicationFn
		serializer           persistence.PayloadSerializer
		replicationTimeout   time.Duration
		logger               log.Logger
	}
)

// NewNDCHistoryRereplicator create a new NDCHistoryRereplicatorImpl
func NewNDCHistoryRereplicator(
	targetClusterName string,
	domainCache cache.DomainCache,
	adminClient a.Client,
	historyReplicationFn nDCHistoryReplicationFn,
	serializer persistence.PayloadSerializer,
	replicationTimeout time.Duration,
	logger log.Logger,
) *NDCHistoryRereplicatorImpl {

	return &NDCHistoryRereplicatorImpl{
		targetClusterName:    targetClusterName,
		domainCache:          domainCache,
		adminClient:          adminClient,
		historyReplicationFn: historyReplicationFn,
		serializer:           serializer,
		replicationTimeout:   replicationTimeout,
		logger:               logger,
	}
}

// SendSingleWorkflowHistory sends one run IDs's history events to remote
func (n *NDCHistoryRereplicatorImpl) SendSingleWorkflowHistory(
	domainID string,
	workflowID string,
	runID string,
	startEventID *int64,
	startEventVersion *int64,
	endEventID *int64,
	endEventVersion *int64,
) error {

	var token []byte
	for doPaging := true; doPaging; doPaging = len(token) > 0 {
		response, err := n.getHistory(
			domainID,
			workflowID,
			runID,
			startEventID,
			startEventVersion,
			endEventID,
			endEventVersion,
			token,
			defaultPageSize)
		if err != nil {
			return err
		}
		token = response.NextPageToken

		// we don't need to handle continue as new here
		// because we only care about the current run
		// TODO: revisit to evaluate if we need to handle continue as new
		for _, batch := range response.HistoryBatches {
			replicationRequest := n.createReplicationRawRequest(
				domainID,
				workflowID,
				runID,
				batch,
				response.GetVersionHistory().GetItems())
			err := n.sendReplicationRawRequest(replicationRequest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *NDCHistoryRereplicatorImpl) createReplicationRawRequest(
	domainID string,
	workflowID string,
	runID string,
	historyBlob *shared.DataBlob,
	versionHistoryItems []*shared.VersionHistoryItem,
) *history.ReplicateEventsV2Request {

	request := &history.ReplicateEventsV2Request{
		DomainUUID: common.StringPtr(domainID),
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		Events:              historyBlob,
		VersionHistoryItems: versionHistoryItems,
	}

	return request
}

func (n *NDCHistoryRereplicatorImpl) sendReplicationRawRequest(
	request *history.ReplicateEventsV2Request,
) error {

	if request == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), n.replicationTimeout)
	defer func() {
		cancel()
	}()
	return n.historyReplicationFn(ctx, request)
}

func (n *NDCHistoryRereplicatorImpl) getHistory(
	domainID string,
	workflowID string,
	runID string,
	startEventID *int64,
	startEventVersion *int64,
	endEventID *int64,
	endEventVersion *int64,
	token []byte,
	pageSize int32,
) (*admin.GetWorkflowExecutionRawHistoryV2Response, error) {

	logger := n.logger.WithTags(tag.WorkflowRunID(runID))

	domainEntry, err := n.domainCache.GetDomainByID(domainID)
	if err != nil {
		logger.Error("error getting domain", tag.Error(err))
		return nil, err
	}
	domainName := domainEntry.GetInfo().Name

	ctx, cancel := context.WithTimeout(context.Background(), n.replicationTimeout)
	defer cancel()
	response, err := n.adminClient.GetWorkflowExecutionRawHistoryV2(ctx, &admin.GetWorkflowExecutionRawHistoryV2Request{
		Domain: common.StringPtr(domainName),
		Execution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		StartEventId:      startEventID,
		StartEventVersion: startEventVersion,
		EndEventId:        endEventID,
		EndEventVersion:   endEventVersion,
		MaximumPageSize:   common.Int32Ptr(pageSize),
		NextPageToken:     token,
	})
	if err != nil {
		logger.Error("error getting history", tag.Error(err))
		return nil, err
	}

	return response, nil
}

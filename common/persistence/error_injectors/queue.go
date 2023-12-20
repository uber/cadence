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

type queueClient struct {
	persistence persistence.QueueManager
	errorRate   float64
	logger      log.Logger
}

// NewQueueClient creates an error injection client to manage queue
func NewQueueClient(
	persistence persistence.QueueManager,
	errorRate float64,
	logger log.Logger,
) persistence.QueueManager {
	return &queueClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *queueClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.EnqueueMessage(ctx, message)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationEnqueueMessage,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*persistence.QueueMessage, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response []*persistence.QueueMessage
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateAckLevel,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response map[string]int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetAckLevels(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetAckLevels,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteMessagesBefore(ctx, messageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteMessagesBefore,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.EnqueueMessageToDLQ(ctx, message)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationEnqueueMessageToDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*persistence.QueueMessage, []byte, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response []*persistence.QueueMessage
	var token []byte
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, token, persistenceErr = p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadMessagesFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, nil, fakeErr
	}
	return response, token, persistenceErr
}

func (p *queueClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeDeleteMessagesFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateDLQAckLevel,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response map[string]int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDLQAckLevels(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDLQAckLevels,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDLQSize(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDLQSize,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return 0, fakeErr
	}
	return response, persistenceErr
}

func (p *queueClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteMessageFromDLQ(ctx, messageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteMessageFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueClient) Close() {
	p.persistence.Close()
}

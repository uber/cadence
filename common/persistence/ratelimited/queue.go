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

type queueClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.QueueManager
	logger      log.Logger
}

// NewQueueClient creates a client to manage queue
func NewQueueClient(
	persistence persistence.QueueManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.QueueManager {
	return &queueClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *queueClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.EnqueueMessage(ctx, message)
}

func (p *queueClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*persistence.QueueMessage, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
}

func (p *queueClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
}

func (p *queueClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetAckLevels(ctx)
}

func (p *queueClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteMessagesBefore(ctx, messageID)
}

func (p *queueClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.EnqueueMessageToDLQ(ctx, message)
}

func (p *queueClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*persistence.QueueMessage, []byte, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
}

func (p *queueClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
}

func (p *queueClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
}

func (p *queueClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetDLQAckLevels(ctx)
}

func (p *queueClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return 0, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetDLQSize(ctx)
}

func (p *queueClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteMessageFromDLQ(ctx, messageID)
}

func (p *queueClient) Close() {
	p.persistence.Close()
}

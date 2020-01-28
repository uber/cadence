// Copyright (c) 2020 Uber Technologies, Inc.
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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination dlqTaskHandler_mock.go

package domain

import (
	"fmt"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	// DLQTaskHandler is the interface handles domain DLQ messages
	DLQTaskHandler interface {
		ReadMessages(lastMessageID int, pageSize int, pageToken []byte) ([]*replicator.ReplicationTask, []byte, error)
		PurgeMessages(lastMessageID int) error
		MergeMessages(lastMessageID int, pageSize int, pageToken []byte) ([]byte, error)
	}

	dlqTaskHandlerImpl struct {
		replicationHandler     ReplicationHandler
		domainReplicationQueue persistence.DomainReplicationQueue
		logger                 log.Logger
	}
)

// NewDLQTaskHandler returns a DLQTaskHandler instance
func NewDLQTaskHandler(
	replicationHandler ReplicationHandler,
	domainReplicationQueue persistence.DomainReplicationQueue,
	logger log.Logger,
) DLQTaskHandler {
	return &dlqTaskHandlerImpl{
		replicationHandler:     replicationHandler,
		domainReplicationQueue: domainReplicationQueue,
		logger:                 logger,
	}
}

// ReadMessages reads domain replication DLQ messages
func (d *dlqTaskHandlerImpl) ReadMessages(
	lastMessageID int,
	pageSize int,
	pageToken []byte,
) ([]*replicator.ReplicationTask, []byte, error) {

	ackLevel, err := d.domainReplicationQueue.GetDLQAckLevel()
	if err != nil {
		return nil, nil, err
	}

	return d.domainReplicationQueue.GetMessagesFromDLQ(
		ackLevel,
		lastMessageID,
		pageSize,
		pageToken,
	)
}

// PurgeMessages purges domain replication DLQ messages
func (d *dlqTaskHandlerImpl) PurgeMessages(
	lastMessageID int,
) error {

	ackLevel, err := d.domainReplicationQueue.GetDLQAckLevel()
	if err != nil {
		return err
	}

	if err := d.domainReplicationQueue.RangeDeleteMessagesFromDLQ(
		ackLevel,
		lastMessageID,
	); err != nil {
		return err
	}

	if err := d.domainReplicationQueue.UpdateDLQAckLevel(
		lastMessageID,
	); err != nil {
		d.logger.Error("Failed to update DLQ ack level after purging messages", tag.Error(err))
	}

	return nil
}

// MergeMessages merges domain replication DLQ messages
func (d *dlqTaskHandlerImpl) MergeMessages(
	lastMessageID int,
	pageSize int,
	pageToken []byte,
) ([]byte, error) {

	ackLevel, err := d.domainReplicationQueue.GetDLQAckLevel()
	if err != nil {
		return nil, err
	}

	messages, token, err := d.domainReplicationQueue.GetMessagesFromDLQ(
		ackLevel,
		lastMessageID,
		pageSize,
		pageToken,
	)
	if err != nil {
		return nil, err
	}

	var ackedMessageID int
	for _, message := range messages {
		domainTask := message.GetDomainTaskAttributes()
		if domainTask == nil {
			return nil, &shared.InternalServiceError{Message: "Encounter non domain replication task in domain replication queue."}
		}

		if err := d.replicationHandler.HandleReceivingTask(
			domainTask,
		); err != nil {
			return nil, err
		}

		if err := d.domainReplicationQueue.DeleteMessageFromDLQ(int(*message.SourceTaskId)); err != nil {
			return nil, &shared.InternalServiceError{
				Message: fmt.Sprintf("Failed to delete DLQ message after merge. Last Merged Message ID: %v, Error: %v.",
					ackedMessageID,
					err,
				),
			}
		}
		ackedMessageID = int(*message.SourceTaskId)
	}

	if err := d.domainReplicationQueue.UpdateDLQAckLevel(ackedMessageID); err != nil {
		d.logger.Error("failed to update ack level on merging domain DLQ message", tag.Error(err))
	}

	return token, nil
}

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

package cassandra

import (
	"context"
	"fmt"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/service/config"
)

const (
	emptyMessageID = -1
)

type (
	nosqlQueueManager struct {
		queueType persistence.QueueType
		logger    log.Logger
		db        nosqlplugin.DB
	}
)

func (q *nosqlQueueManager) Close() {
	q.db.Close()
}

func newQueue(
	cfg config.Cassandra,
	logger log.Logger,
	queueType persistence.QueueType,
) (persistence.Queue, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg)
	if err != nil {
		return nil, err
	}

	// TODO Is that not needed anymore?
	//retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	//retryPolicy.SetBackoffCoefficient(1.5)
	//retryPolicy.SetMaximumAttempts(5)

	queue := &nosqlQueueManager{
		db:        db,
		logger:    logger,
		queueType: queueType,
	}
	if err := queue.createQueueMetadataEntryIfNotExist(); err != nil {
		return nil, fmt.Errorf("failed to check and create queue metadata entry: %v", err)
	}

	return queue, nil
}

func (q *nosqlQueueManager) createQueueMetadataEntryIfNotExist() error {
	queueMetadata, err := q.getQueueMetadata(context.Background(), q.queueType)
	if err != nil {
		return err
	}

	if queueMetadata == nil {
		if err := q.insertInitialQueueMetadataRecord(context.Background(), q.queueType); err != nil {
			return err
		}
	}

	dlqMetadata, err := q.getQueueMetadata(context.Background(), q.getDLQTypeFromQueueType())
	if err != nil {
		return err
	}

	if dlqMetadata == nil {
		return q.insertInitialQueueMetadataRecord(context.Background(), q.getDLQTypeFromQueueType())
	}

	return nil
}

func (q *nosqlQueueManager) EnqueueMessage(
	ctx context.Context,
	messagePayload []byte,
) error {
	lastMessageID, err := q.getLastMessageID(ctx, q.queueType)
	if err != nil {
		return err
	}

	_, err = q.tryEnqueue(ctx, q.queueType, lastMessageID+1, messagePayload)
	return err
}

func (q *nosqlQueueManager) EnqueueMessageToDLQ(
	ctx context.Context,
	messagePayload []byte,
) (int64, error) {
	// Use negative queue type as the dlq type
	lastMessageID, err := q.getLastMessageID(ctx, q.getDLQTypeFromQueueType())
	if err != nil {
		return emptyMessageID, err
	}

	// Use negative queue type as the dlq type
	return q.tryEnqueue(ctx, q.getDLQTypeFromQueueType(), lastMessageID+1, messagePayload)
}

func (q *nosqlQueueManager) tryEnqueue(
	ctx context.Context,
	queueType persistence.QueueType,
	messageID int64,
	messagePayload []byte,
) (int64, error) {
	err := q.db.InsertIntoQueue(ctx, &nosqlplugin.QueueMessageRow{
		QueueType: queueType,
		ID:        messageID,
		Payload:   messagePayload,
	})
	if err != nil {
		if q.db.IsThrottlingError(err) {
			return emptyMessageID, &shared.ServiceBusyError{
				Message: fmt.Sprintf("Failed to enqueue message. Error: %v, Type: %v.", err, queueType),
			}
		}
		if q.db.IsConditionFailedError(err) {
			return emptyMessageID, &persistence.ConditionFailedError{Msg: fmt.Sprintf("message ID %v exists in queue", messageID)}
		}
		return emptyMessageID, &shared.InternalServiceError{
			Message: fmt.Sprintf("Failed to enqueue message. Error: %v, Type: %v.", err, queueType),
		}
	}

	return messageID, nil
}

func (q *nosqlQueueManager) getLastMessageID(
	ctx context.Context,
	queueType persistence.QueueType,
) (int64, error) {

	msgID, err := q.db.GetLastEnqueuedMessageID(ctx, queueType)
	if err != nil {
		if q.db.IsNotFoundError(err) {
			return emptyMessageID, nil
		} else if q.db.IsThrottlingError(err) {
			return emptyMessageID, &shared.ServiceBusyError{
				Message: fmt.Sprintf("Failed to get last message ID for queue %v. Error: %v", queueType, err),
			}
		}

		return emptyMessageID, &shared.InternalServiceError{
			Message: fmt.Sprintf("Failed to get last message ID for queue %v. Error: %v", queueType, err),
		}
	}

	return msgID, nil
}

func (q *nosqlQueueManager) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*persistence.QueueMessage, error) {
	messages, err := q.db.GetMessagesFromQueue(ctx, q.queueType, lastMessageID, maxCount)
	if err != nil {
		return nil, err
	}
	var result []*persistence.QueueMessage
	for _, msg := range messages {
		result = append(result, &persistence.QueueMessage{
			ID:        msg.ID,
			QueueType: msg.QueueType,
			Payload:   msg.Payload,
		})
	}

	return result, nil
}

func (q *nosqlQueueManager) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*persistence.QueueMessage, []byte, error) {
	response, err := q.db.GetMessagesBetween(ctx, nosqlplugin.GetMessagesBetweenRequest{
		QueueType:               q.queueType,
		ExclusiveBeginMessageID: firstMessageID,
		InclusiveEndMessageID:   lastMessageID,
		PageSize:                pageSize,
		NextPageToken:           pageToken,
	})
	if err != nil {
		return nil, nil, err
	}
	var result []*persistence.QueueMessage
	for _, msg := range response.Rows {
		result = append(result, &persistence.QueueMessage{
			ID:        msg.ID,
			QueueType: msg.QueueType,
			Payload:   msg.Payload,
		})
	}

	return result, response.NextPageToken, nil
}

func (q *nosqlQueueManager) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {

	return q.db.DeleteMessagesBefore(ctx, q.queueType, messageID)
}

func (q *nosqlQueueManager) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	// Use negative queue type as the dlq type
	return q.db.DeleteMessage(ctx, q.queueType, messageID)
}

func (q *nosqlQueueManager) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	// Use negative queue type as the dlq type
	return q.db.RangeDeleteMessages(ctx, q.queueType, firstMessageID, lastMessageID)
}

func (q *nosqlQueueManager) insertInitialQueueMetadataRecord(
	ctx context.Context,
	queueType persistence.QueueType,
) error {
	version := int64(0)
	return q.db.InitQueueMetadata(ctx, queueType, version)
}

func (q *nosqlQueueManager) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {

	return q.updateAckLevel(ctx, messageID, clusterName, q.queueType)
}

func (q *nosqlQueueManager) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	queueMetadata, err := q.getQueueMetadata(ctx, q.queueType)
	if err != nil {
		return nil, err
	}

	return queueMetadata.ClusterAckLevels, nil
}

func (q *nosqlQueueManager) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {

	return q.updateAckLevel(ctx, messageID, clusterName, q.getDLQTypeFromQueueType())
}

func (q *nosqlQueueManager) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {

	// Use negative queue type as the dlq type
	queueMetadata, err := q.getQueueMetadata(ctx, q.getDLQTypeFromQueueType())
	if err != nil {
		return nil, err
	}

	return queueMetadata.ClusterAckLevels, nil
}

func (q *nosqlQueueManager) getQueueMetadata(
	ctx context.Context,
	queueType persistence.QueueType,
) (*nosqlplugin.QueueMetadataRow, error) {
	row, err := q.db.GetQueueMetadata(ctx, queueType)
	if err != nil {
		if q.db.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get queue metadata: %v", err)
	}

	return row, nil
}

func (q *nosqlQueueManager) updateQueueMetadata(
	ctx context.Context,
	metadata *nosqlplugin.QueueMetadataRow,
	queueType persistence.QueueType,
) error {
	err := q.db.UpdateQueueMetadataCas(ctx, *metadata)
	if err != nil {
		if q.db.IsConditionFailedError(err) {
			return &shared.InternalServiceError{
				Message: fmt.Sprintf("UpdateAckLevel operation encounter concurrent write."),
			}
		}
		return &shared.InternalServiceError{
			Message: fmt.Sprintf("UpdateAckLevel operation failed. Error %v", err),
		}
	}

	return nil
}

func (q *nosqlQueueManager) getDLQTypeFromQueueType() persistence.QueueType {
	return -q.queueType
}

func (q *nosqlQueueManager) updateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
	queueType persistence.QueueType,
) error {

	queueMetadata, err := q.getQueueMetadata(ctx, queueType)
	if err != nil {
		return &shared.InternalServiceError{
			Message: fmt.Sprintf("UpdateDLQAckLevel operation failed. Error %v", err),
		}
	}

	// Ignore possibly delayed message
	if queueMetadata.ClusterAckLevels[clusterName] > messageID {
		return nil
	}

	queueMetadata.ClusterAckLevels[clusterName] = messageID
	queueMetadata.Version++

	// Use negative queue type as the dlq type
	err = q.updateQueueMetadata(ctx, queueMetadata, queueType)
	if err != nil {
		return &shared.InternalServiceError{
			Message: fmt.Sprintf("UpdateDLQAckLevel operation failed. Error %v", err),
		}
	}
	return nil
}

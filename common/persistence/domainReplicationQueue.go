package persistence

import (
	"errors"
	"fmt"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common/codec"
)

var _ DomainReplicationQueue = (*domainReplicationQueueImpl)(nil)

func NewDomainReplicationQueue(queue Queue) DomainReplicationQueue {
	return &domainReplicationQueueImpl{
		queue:   queue,
		encoder: codec.NewThriftRWEncoder(),
	}
}

type (
	domainReplicationQueueImpl struct {
		queue   Queue
		encoder codec.BinaryEncoder
	}
)

func (q *domainReplicationQueueImpl) Publish(message interface{}) error {
	task, ok := message.(*replicator.ReplicationTask)
	if !ok {
		return errors.New("wrong message type")
	}

	bytes, err := q.encoder.Encode(task)
	if err != nil {
		return fmt.Errorf("failed to encode message: %v", err)
	}
	return q.queue.EnqueueMessage(bytes)
}

func (q *domainReplicationQueueImpl) GetReplicationMessages(
	lastMessageID int,
	maxCount int,
) ([]*replicator.ReplicationTask, int, error) {
	messages, err := q.queue.GetMessages(lastMessageID, maxCount)
	if err != nil {
		return nil, lastMessageID, err
	}

	var replicationTasks []*replicator.ReplicationTask
	for _, message := range messages {
		var replicationTask replicator.ReplicationTask
		err := q.encoder.Decode(message.Payload, &replicationTask)
		if err != nil {
			return nil, lastMessageID, fmt.Errorf("failed to decode task: %v", err)
		}

		lastMessageID = message.ID
		replicationTasks = append(replicationTasks, &replicationTask)
	}

	return replicationTasks, lastMessageID, nil
}

package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/queue"
	"github.com/uber/cadence/common/service/config"
)

const (
	templateInsertQueueQuery      = `INSERT INTO queue (queue_type, message_id, message) VALUES(?, ?, ?) IF NOT EXISTS`
	templateGetLastMessageIDQuery = `SELECT message_id FROM queue WHERE queue_type=? ORDER BY message_id DESC LIMIT 1`
)

type (
	cassandraQueue struct {
		queueType      string
		messageEncoder messaging.MessageEncoder
		retryPolicy    backoff.RetryPolicy
		cassandraStore
	}

	messageIDConflictError struct {
		lastMessageID int
	}
)

func NewQueue(
	cfg config.Cassandra,
	logger log.Logger,
	queueType string,
	messageEncoder messaging.MessageEncoder,
) (queue.Queue, error) {
	cluster := NewCassandraCluster(cfg.Hosts, cfg.Port, cfg.User, cfg.Password, cfg.Datacenter)
	cluster.Keyspace = cfg.Keyspace
	cluster.ProtoVersion = cassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = defaultSessionTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	retryPolicy := backoff.NewExponentialRetryPolicy(50 * time.Millisecond)
	retryPolicy.SetBackoffCoefficient(1.2)
	retryPolicy.SetMaximumAttempts(5)

	return &cassandraQueue{
		cassandraStore: cassandraStore{session: session, logger: logger},
		queueType:      queueType,
		messageEncoder: messageEncoder,
	}, nil
}

func (q *cassandraQueue) Publish(message interface{}) error {
	encodedMessage, err := q.messageEncoder(message)
	if err != nil {
		return err
	}

	var nextMessageID int
	err = backoff.Retry(func() error {
		nextMessageID, err = q.getNextMessageID()
		return err
	}, q.retryPolicy, common.IsPersistenceTransientError)

	if err != nil {
		return fmt.Errorf("failed to get next messageID: %v", err)
	}

	return backoff.Retry(func() error {
		err = q.publish(nextMessageID, encodedMessage)
		if messageIDConflictError, ok := err.(*messageIDConflictError); ok {
			nextMessageID = messageIDConflictError.lastMessageID
		}
		return err
	}, q.retryPolicy, func(e error) bool {
		return common.IsPersistenceTransientError(e) || isMessageIDConflictError(e)
	})
}

func (q *cassandraQueue) publish(messageID int, message []byte) error {
	query := q.session.Query(templateInsertQueueQuery, q.queueType, messageID, message)
	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		if isThrottlingError(err) {
			return &workflow.ServiceBusyError{
				Message: fmt.Sprintf("Failed to enqueue message. Error: %v", err),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to enqueue message. Error: %v", err),
		}
	}

	if !applied {
		return &messageIDConflictError{lastMessageID: previous["message_id"].(int)}
	}

	return nil
}

func (q *cassandraQueue) getNextMessageID() (int, error) {
	query := q.session.Query(templateGetLastMessageIDQuery, q.queueType)

	result := make(map[string]interface{})
	err := query.MapScan(result)
	if err != nil {
		if err == gocql.ErrNotFound {
			return 0, nil
		} else if isThrottlingError(err) {
			return 0, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("Failed to get last message ID for queue %v. Error: %v", q.queueType, err),
			}
		}

		return 0, &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to get last message ID for queue %v. Error: %v", q.queueType, err),
		}
	}

	return result["message_id"].(int) + 1, nil
}

func (q *cassandraQueue) Close() error {
	if q.session != nil {
		q.session.Close()
	}

	return nil
}

func (e *messageIDConflictError) Error() string {
	return fmt.Sprintf("message ID %v exists in queue", e.lastMessageID)
}

func isMessageIDConflictError(err error) bool {
	_, ok := err.(*messageIDConflictError)
	return ok
}

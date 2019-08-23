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

package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

type (
	sqlQueue struct {
		queueType   int
		logger      log.Logger
		retryPolicy backoff.RetryPolicy
		sqlStore
	}
)

func newQueue(
	db sqldb.Interface,
	logger log.Logger,
	queueType int,
) (persistence.Queue, error) {
	retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	retryPolicy.SetBackoffCoefficient(1.5)
	retryPolicy.SetMaximumAttempts(10)

	return &sqlQueue{
		sqlStore: sqlStore{
			db:     db,
			logger: logger,
		},
		queueType:   queueType,
		logger:      logger,
		retryPolicy: retryPolicy,
	}, nil

}

func (q *sqlQueue) EnqueueMessage(messagePayload []byte) error {
	err := backoff.Retry(func() error {
		return q.tryEnqueue(messagePayload)
	}, q.retryPolicy,
		func(e error) bool {
			_, ok := e.(*persistence.ConditionFailedError)
			return ok
		})
	if err != nil {
		return &workflow.InternalServiceError{Message: err.Error()}
	}
	return nil
}

func (q *sqlQueue) tryEnqueue(messagePayload []byte) error {
	return q.txExecute("EnqueueMessage", func(tx sqldb.Tx) error {
		lastMessageID, err := tx.GetLastEnqueuedMessageID(q.queueType)
		if err != nil {
			if err == sql.ErrNoRows {
				lastMessageID = -1
			} else {
				return fmt.Errorf("failed to get last enqueued message id: %v", err)
			}
		}

		_, err = tx.InsertIntoQueue(&sqldb.QueueRow{QueueType: q.queueType, MessageID: lastMessageID + 1, MessagePayload: messagePayload})
		if err != nil {
			if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number == ErrDupEntry {
				return &persistence.ConditionFailedError{Msg: err.Error()}
			}
			return err
		}
		return nil
	})
}

func (q *sqlQueue) GetMessages(lastMessageID, maxCount int) ([]*persistence.QueueMessage, error) {
	rows, err := q.db.GetMessagesFromQueue(q.queueType, lastMessageID, maxCount)
	if err != nil {
		return nil, err
	}

	var messages []*persistence.QueueMessage
	for _, row := range rows {
		messages = append(messages, &persistence.QueueMessage{ID: row.MessageID, Payload: row.MessagePayload})
	}
	return messages, nil
}

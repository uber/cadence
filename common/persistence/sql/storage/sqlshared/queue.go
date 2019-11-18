// Copyright (c) 2017 Uber Technologies, Inc.
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

package sqlshared

import (
	"database/sql"
	"encoding/json"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// InsertIntoQueue inserts a new row into queue table
func (mdb *DB) InsertIntoQueue(row *sqldb.QueueRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.EnqueueMessageQuery(), row)
}

// GetLastEnqueuedMessageIDForUpdate returns the last enqueued message ID
func (mdb *DB) GetLastEnqueuedMessageIDForUpdate(queueType common.QueueType) (int, error) {
	var lastMessageID int
	err := mdb.conn.Get(&lastMessageID, mdb.driver.GetLastMessageIDQuery(), queueType)
	return lastMessageID, err
}

// GetMessagesFromQueue retrieves messages from the queue
func (mdb *DB) GetMessagesFromQueue(queueType common.QueueType, lastMessageID, maxRows int) ([]sqldb.QueueRow, error) {
	var rows []sqldb.QueueRow
	err := mdb.conn.Select(&rows, mdb.driver.GetMessagesQuery(), queueType, lastMessageID, maxRows)
	return rows, err
}

// DeleteMessagesBefore deletes messages before messageID from the queue
func (mdb *DB) DeleteMessagesBefore(queueType common.QueueType, messageID int) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteMessagesQuery(), queueType, messageID)
}

// InsertAckLevel inserts ack level
func (mdb *DB) InsertAckLevel(queueType common.QueueType, messageID int, clusterName string) error {
	clusterAckLevels := map[string]int{clusterName: messageID}
	data, err := json.Marshal(clusterAckLevels)
	if err != nil {
		return err
	}

	_, err = mdb.conn.NamedExec(mdb.driver.InsertQueueMetadataQuery(), sqldb.QueueMetadataRow{QueueType: queueType, Data: data})
	return err

}

// UpdateAckLevels updates cluster ack levels
func (mdb *DB) UpdateAckLevels(queueType common.QueueType, clusterAckLevels map[string]int) error {
	data, err := json.Marshal(clusterAckLevels)
	if err != nil {
		return err
	}

	_, err = mdb.conn.Exec(mdb.driver.UpdateQueueMetadataQuery(), data, queueType)
	return err
}

// GetAckLevels returns ack levels for pulling clusters
func (mdb *DB) GetAckLevels(queueType common.QueueType, forUpdate bool) (map[string]int, error) {
	queryStr := mdb.driver.GetQueueMetadataQuery()
	if forUpdate {
		queryStr = mdb.driver.GetQueueMetadataForUpdateQuery()
	}

	var data []byte
	err := mdb.conn.Get(&data, queryStr, queueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	var clusterAckLevels map[string]int
	if err := json.Unmarshal(data, &clusterAckLevels); err != nil {
		return nil, err
	}

	return clusterAckLevels, nil
}

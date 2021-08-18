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

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

const (
	templateEnqueueMessageQuery            = `INSERT INTO queue (queue_type, message_id, message_payload) VALUES(:queue_type, :message_id, :message_payload)`
	templateGetLastMessageIDQuery          = `SELECT message_id FROM queue WHERE queue_type=$1 ORDER BY message_id DESC LIMIT 1 FOR UPDATE`
	templateGetMessagesQuery               = `SELECT message_id, message_payload FROM queue WHERE queue_type = $1 and message_id > $2 ORDER BY message_id ASC LIMIT $3`
	templateGetMessagesBetweenQuery        = `SELECT message_id, message_payload FROM queue WHERE queue_type = $1 and messageid > $2 and message_id <= $3 ORDER BY message_id ASC LIMIT $4`
	templateDeleteMessageQuery             = `DELETE FROM queue WHERE queue_type = $1 and message_id = $2`
	templateDeleteMessagesBeforeQuery      = `DELETE FROM queue WHERE queue_type = $1 and message_id < $2`
	templateRangeDeleteMessagesQuery       = `DELETE FROM queue WHERE queue_type = $1 and message_id > $2 and message_id <= $3`
	templateGetQueueMetadataQuery          = `SELECT data from queue_metadata WHERE queue_type = $1`
	templateGetQueueMetadataForUpdateQuery = templateGetQueueMetadataQuery + ` FOR UPDATE`
	templateInsertQueueMetadataQuery       = `INSERT INTO queue_metadata (queue_type, data) VALUES(:queue_type, :data)`
	templateUpdateQueueMetadataQuery       = `UPDATE queue_metadata SET data = $1 WHERE queue_type = $2`
	templateGetQueueSizeQuery              = `SELECT COUNT(1) AS count FROM queue WHERE queue_type=$1`
)

// InsertIntoQueue inserts a new row into queue table
func (pdb *db) InsertIntoQueue(ctx context.Context, row *sqlplugin.QueueRow) (sql.Result, error) {
	return pdb.conn.NamedExecContext(ctx, templateEnqueueMessageQuery, row)
}

// GetLastEnqueuedMessageIDForUpdate returns the last enqueued message ID
func (pdb *db) GetLastEnqueuedMessageIDForUpdate(ctx context.Context, queueType persistence.QueueType) (int64, error) {
	var lastMessageID int64
	err := pdb.conn.GetContext(ctx, &lastMessageID, templateGetLastMessageIDQuery, queueType)
	return lastMessageID, err
}

// GetMessagesFromQueue retrieves messages from the queue
func (pdb *db) GetMessagesFromQueue(ctx context.Context, queueType persistence.QueueType, lastMessageID int64, maxRows int) ([]sqlplugin.QueueRow, error) {
	var rows []sqlplugin.QueueRow
	err := pdb.conn.SelectContext(ctx, &rows, templateGetMessagesQuery, queueType, lastMessageID, maxRows)
	return rows, err
}

// GetMessagesBetween retrieves messages from the queue
func (pdb *db) GetMessagesBetween(ctx context.Context, queueType persistence.QueueType, firstMessageID int64, lastMessageID int64, maxRows int) ([]sqlplugin.QueueRow, error) {
	var rows []sqlplugin.QueueRow
	err := pdb.conn.SelectContext(ctx, &rows, templateGetMessagesBetweenQuery, queueType, firstMessageID, lastMessageID, maxRows)
	return rows, err
}

// DeleteMessagesBefore deletes messages before messageID from the queue
func (pdb *db) DeleteMessagesBefore(ctx context.Context, queueType persistence.QueueType, messageID int64) (sql.Result, error) {
	return pdb.conn.ExecContext(ctx, templateDeleteMessagesBeforeQuery, queueType, messageID)
}

// RangeDeleteMessages deletes messages before messageID from the queue
func (pdb *db) RangeDeleteMessages(ctx context.Context, queueType persistence.QueueType, exclusiveBeginMessageID int64, inclusiveEndMessageID int64) (sql.Result, error) {
	return pdb.conn.ExecContext(ctx, templateRangeDeleteMessagesQuery, queueType, exclusiveBeginMessageID, inclusiveEndMessageID)
}

// DeleteMessage deletes message with a messageID from the queue
func (pdb *db) DeleteMessage(ctx context.Context, queueType persistence.QueueType, messageID int64) (sql.Result, error) {
	return pdb.conn.ExecContext(ctx, templateDeleteMessageQuery, queueType, messageID)
}

// InsertAckLevel inserts ack level
func (pdb *db) InsertAckLevel(ctx context.Context, queueType persistence.QueueType, messageID int64, clusterName string) error {
	clusterAckLevels := map[string]int64{clusterName: messageID}
	data, err := json.Marshal(clusterAckLevels)
	if err != nil {
		return err
	}

	_, err = pdb.conn.NamedExecContext(ctx, templateInsertQueueMetadataQuery, sqlplugin.QueueMetadataRow{QueueType: queueType, Data: data})
	return err

}

// UpdateAckLevels updates cluster ack levels
func (pdb *db) UpdateAckLevels(ctx context.Context, queueType persistence.QueueType, clusterAckLevels map[string]int64) error {
	data, err := json.Marshal(clusterAckLevels)
	if err != nil {
		return err
	}

	_, err = pdb.conn.ExecContext(ctx, templateUpdateQueueMetadataQuery, data, queueType)
	return err
}

// GetAckLevels returns ack levels for pulling clusters
func (pdb *db) GetAckLevels(ctx context.Context, queueType persistence.QueueType, forUpdate bool) (map[string]int64, error) {
	queryStr := templateGetQueueMetadataQuery
	if forUpdate {
		queryStr = templateGetQueueMetadataForUpdateQuery
	}

	var data []byte
	err := pdb.conn.GetContext(ctx, &data, queryStr, queueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	var clusterAckLevels map[string]int64
	if err := json.Unmarshal(data, &clusterAckLevels); err != nil {
		return nil, err
	}

	return clusterAckLevels, nil
}

// GetQueueSize returns the queue size
func (pdb *db) GetQueueSize(
	ctx context.Context,
	queueType persistence.QueueType,
) (int64, error) {

	var size []int64
	if err := pdb.conn.SelectContext(
		ctx,
		&size,
		templateGetQueueSizeQuery,
		queueType,
	); err != nil {
		return 0, err
	}
	return size[0], nil
}

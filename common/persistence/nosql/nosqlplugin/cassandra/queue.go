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

package cassandra

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

const (
	templateEnqueueMessageQuery             = `INSERT INTO queue (queue_type, message_id, message_payload) VALUES(?, ?, ?) IF NOT EXISTS`
	templateGetLastMessageIDQuery           = `SELECT message_id FROM queue WHERE queue_type=? ORDER BY message_id DESC LIMIT 1`
	templateGetMessagesQuery                = `SELECT message_id, message_payload FROM queue WHERE queue_type = ? and message_id > ? LIMIT ?`
	templateGetMessagesFromDLQQuery         = `SELECT message_id, message_payload FROM queue WHERE queue_type = ? and message_id > ? and message_id <= ?`
	templateRangeDeleteMessagesBeforeQuery  = `DELETE FROM queue WHERE queue_type = ? and message_id < ?`
	templateRangeDeleteMessagesBetweenQuery = `DELETE FROM queue WHERE queue_type = ? and message_id > ? and message_id <= ?`
	templateDeleteMessageQuery              = `DELETE FROM queue WHERE queue_type = ? and message_id = ?`
	templateGetQueueMetadataQuery           = `SELECT cluster_ack_level, version FROM queue_metadata WHERE queue_type = ?`
	templateInsertQueueMetadataQuery        = `INSERT INTO queue_metadata (queue_type, cluster_ack_level, version) VALUES(?, ?, ?) IF NOT EXISTS`
	templateUpdateQueueMetadataQuery        = `UPDATE queue_metadata SET cluster_ack_level = ?, version = ? WHERE queue_type = ? IF version = ?`
	templateGetQueueSizeQuery               = `SELECT COUNT(1) AS count FROM queue WHERE queue_type=?`
)

//Insert message into queue, return error if failed or already exists
// Must return ConditionFailure error if row already exists
func (db *cdb) InsertIntoQueue(
	ctx context.Context,
	row *nosqlplugin.QueueMessageRow,
) error {
	query := db.session.Query(templateEnqueueMessageQuery, row.QueueType, row.ID, row.Payload).WithContext(ctx)
	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return err
	}

	if !applied {
		return nosqlplugin.NewConditionFailure("queue")
	}
	return nil
}

// Get the ID of last message inserted into the queue
func (db *cdb) SelectLastEnqueuedMessageID(
	ctx context.Context,
	queueType persistence.QueueType,
) (int64, error) {
	query := db.session.Query(templateGetLastMessageIDQuery, queueType).WithContext(ctx)
	result := make(map[string]interface{})
	err := query.MapScan(result)
	if err != nil {
		return 0, err
	}

	return result["message_id"].(int64), nil
}

// Read queue messages starting from the exclusiveBeginMessageID
func (db *cdb) SelectMessagesFrom(
	ctx context.Context,
	queueType persistence.QueueType,
	exclusiveBeginMessageID int64,
	maxRows int,
) ([]*nosqlplugin.QueueMessageRow, error) {
	// Reading replication tasks need to be quorum level consistent, otherwise we could loose task
	query := db.session.Query(templateGetMessagesQuery,
		queueType,
		exclusiveBeginMessageID,
		maxRows,
	).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, fmt.Errorf("SelectMessagesFrom operation failed. Not able to create query iterator")
	}

	var result []*nosqlplugin.QueueMessageRow
	message := make(map[string]interface{})
	for iter.MapScan(message) {
		payload := getMessagePayload(message)
		id := getMessageID(message)
		result = append(result, &nosqlplugin.QueueMessageRow{ID: id, Payload: payload})
		message = make(map[string]interface{})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return result, nil
}

// Read queue message starting from exclusiveBeginMessageID int64, inclusiveEndMessageID int64
func (db *cdb) SelectMessagesBetween(
	ctx context.Context,
	request nosqlplugin.SelectMessagesBetweenRequest,
) (*nosqlplugin.SelectMessagesBetweenResponse, error) {
	// Reading replication tasks need to be quorum level consistent, otherwise we could loose task
	// Use negative queue type as the dlq type
	query := db.session.Query(templateGetMessagesFromDLQQuery,
		request.QueueType,
		request.ExclusiveBeginMessageID,
		request.InclusiveEndMessageID,
	).PageSize(request.PageSize).PageState(request.NextPageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, fmt.Errorf("SelectMessagesBetween operation failed. Not able to create query iterator")
	}

	var rows []nosqlplugin.QueueMessageRow
	message := make(map[string]interface{})
	for iter.MapScan(message) {
		payload := getMessagePayload(message)
		id := getMessageID(message)
		rows = append(rows, nosqlplugin.QueueMessageRow{ID: id, Payload: payload})
		message = make(map[string]interface{})
	}

	nextPageToken := iter.PageState()
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return &nosqlplugin.SelectMessagesBetweenResponse{
		Rows:          rows,
		NextPageToken: nextPageToken,
	}, nil
}

// Delete all messages before exclusiveBeginMessageID
func (db *cdb) DeleteMessagesBefore(
	ctx context.Context,
	queueType persistence.QueueType,
	exclusiveBeginMessageID int64,
) error {
	query := db.session.Query(templateRangeDeleteMessagesBeforeQuery, queueType, exclusiveBeginMessageID).WithContext(ctx)
	return query.Exec()
}

// Delete all messages in a range between exclusiveBeginMessageID and inclusiveEndMessageID
func (db *cdb) DeleteMessagesInRange(
	ctx context.Context,
	queueType persistence.QueueType,
	exclusiveBeginMessageID int64,
	inclusiveEndMessageID int64,
) error {
	query := db.session.Query(templateRangeDeleteMessagesBetweenQuery, queueType, exclusiveBeginMessageID, inclusiveEndMessageID).WithContext(ctx)
	return query.Exec()
}

// Delete one message
func (db *cdb) DeleteMessage(
	ctx context.Context,
	queueType persistence.QueueType,
	messageID int64,
) error {
	query := db.session.Query(templateDeleteMessageQuery, queueType, messageID).WithContext(ctx)
	return query.Exec()
}

// Insert an empty metadata row, starting from a version
func (db *cdb) InsertQueueMetadata(
	ctx context.Context,
	queueType persistence.QueueType,
	version int64,
) error {
	clusterAckLevels := map[string]int64{}
	query := db.session.Query(templateInsertQueueMetadataQuery, queueType, clusterAckLevels, version).WithContext(ctx)

	// NOTE: Must pass nils to be compatible with ScyllaDB's LWT behavior
	// "Scylla always returns the old version of the row, regardless of whether the condition is true or not."
	// See also https://docs.scylladb.com/kb/lwt-differences/
	_, err := query.ScanCAS(nil, nil, nil)
	if err != nil {
		return err
	}
	// it's ok if the query is not applied, which means that the record exists already.
	return nil
}

// **Conditionally** update a queue metadata row, if current version is matched(meaning current == row.Version - 1),
// then the current version will increase by one when updating the metadata row
// it should return ConditionFailure if the condition is not met
func (db *cdb) UpdateQueueMetadataCas(
	ctx context.Context,
	row nosqlplugin.QueueMetadataRow,
) error {
	query := db.session.Query(templateUpdateQueueMetadataQuery,
		row.ClusterAckLevels,
		row.Version,
		row.QueueType,
		row.Version-1,
	).WithContext(ctx)

	// NOTE: Must pass nils to be compatible with ScyllaDB's LWT behavior
	// "Scylla always returns the old version of the row, regardless of whether the condition is true or not."
	// See also https://docs.scylladb.com/kb/lwt-differences/
	applied, err := query.ScanCAS(nil, nil, nil, nil)
	if err != nil {
		return err
	}
	if !applied {
		return nosqlplugin.NewConditionFailure("queue")
	}

	return nil
}

// Read a QueueMetadata
func (db *cdb) SelectQueueMetadata(
	ctx context.Context,
	queueType persistence.QueueType,
) (*nosqlplugin.QueueMetadataRow, error) {
	query := db.session.Query(templateGetQueueMetadataQuery, queueType).WithContext(ctx)
	var ackLevels map[string]int64
	var version int64
	err := query.Scan(&ackLevels, &version)
	if err != nil {
		return nil, err
	}

	// if record exist but ackLevels is empty, we initialize the map
	if ackLevels == nil {
		ackLevels = make(map[string]int64)
	}
	return &nosqlplugin.QueueMetadataRow{
		QueueType:        queueType,
		ClusterAckLevels: ackLevels,
		Version:          version,
	}, nil
}

func (db *cdb) GetQueueSize(
	ctx context.Context,
	queueType persistence.QueueType,
) (int64, error) {

	query := db.session.Query(templateGetQueueSizeQuery, queueType).WithContext(ctx)
	result := make(map[string]interface{})

	if err := query.MapScan(result); err != nil {
		return 0, err
	}
	return result["count"].(int64), nil
}

func getMessagePayload(
	message map[string]interface{},
) []byte {

	return message["message_payload"].([]byte)
}

func getMessageID(
	message map[string]interface{},
) int64 {

	return message["message_id"].(int64)
}

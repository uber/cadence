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

package mysql

const (
	templateEnqueueMessageQuery            = `INSERT INTO queue (queue_type, message_id, message_payload) VALUES(:queue_type, :message_id, :message_payload)`
	templateGetLastMessageIDQuery          = `SELECT message_id FROM queue WHERE message_id >= (SELECT message_id FROM queue WHERE queue_type=? ORDER BY message_id DESC LIMIT 1) FOR UPDATE`
	templateGetMessagesQuery               = `SELECT message_id, message_payload FROM queue WHERE queue_type = ? and message_id > ? LIMIT ?`
	templateDeleteMessagesQuery            = `DELETE FROM queue WHERE queue_type = ? and message_id < ?`
	templateGetQueueMetadataQuery          = `SELECT data from queue_metadata WHERE queue_type = ?`
	templateGetQueueMetadataForUpdateQuery = templateGetQueueMetadataQuery + ` FOR UPDATE`
	templateInsertQueueMetadataQuery       = `INSERT INTO queue_metadata (queue_type, data) VALUES(:queue_type, :data)`
	templateUpdateQueueMetadataQuery       = `UPDATE queue_metadata SET data = ? WHERE queue_type = ?`
)

func (d *driver) EnqueueMessageQuery() string {
	return templateEnqueueMessageQuery
}
func (d *driver) GetLastMessageIDQuery() string {
	return templateGetLastMessageIDQuery
}
func (d *driver) GetMessagesQuery() string {
	return templateGetMessagesQuery
}
func (d *driver) DeleteMessagesQuery() string {
	return templateDeleteMessagesQuery
}
func (d *driver) GetQueueMetadataQuery() string {
	return templateGetQueueMetadataQuery
}
func (d *driver) GetQueueMetadataForUpdateQuery() string {
	return templateGetQueueMetadataForUpdateQuery
}
func (d *driver) InsertQueueMetadataQuery() string {
	return templateInsertQueueMetadataQuery
}
func (d *driver) UpdateQueueMetadataQuery() string {
	return templateUpdateQueueMetadataQuery
}

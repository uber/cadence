package mysql

import (
	"database/sql"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

const (
	templateEnqueueMessageQuery   = `INSERT INTO queue (queue_type, message_id, message_payload) VALUES(:queue_type, :message_id, :message_payload)`
	templateGetLastMessageIDQuery = `SELECT message_id FROM queue WHERE queue_type=? ORDER BY message_id DESC LIMIT 1`
	templateGetMessagesQuery      = `SELECT message_id, message_payload FROM queue WHERE queue_type = ? and message_id > ? LIMIT ?`
)

func (mdb *DB) InsertIntoQueue(row *sqldb.QueueRow) (sql.Result, error) {
	return mdb.conn.NamedExec(templateEnqueueMessageQuery, row)
}

func (mdb *DB) GetLastEnqueuedMessageID(queueType int) (int, error) {
	var lastMessageID int
	err := mdb.conn.Get(&lastMessageID, templateGetLastMessageIDQuery, queueType)
	return lastMessageID, err
}

func (mdb *DB) GetMessagesFromQueue(queueType, lastMessageID, maxRows int) ([]sqldb.QueueRow, error) {
	var rows []sqldb.QueueRow
	err := mdb.conn.Select(&rows, templateGetMessagesQuery, queueType, lastMessageID, maxRows)
	return rows, err
}

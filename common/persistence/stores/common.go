package stores

import (
	"github.com/gocql/gocql"
	"time"
)

const (
	CassandraProtoVersion          = 4
	CassandraDefaultSessionTimeout = 10 * time.Second
)

const (
	// Row types for table executions
	CassandraRowTypeShard = iota
	CassandraRowTypeExecution
	CassandraRowTypeTransferTask
	CassandraRowTypeTimerTask
	CassandraRowTypeReplicationTask
	CassandraRowTypeDLQ
)

// Guidelines for creating new special UUID constants
// Each UUID should be of the form: E0000000-R000-f000-f000-00000000000x
// Where x is any hexadecimal value, E represents the entity type valid values are:
// E = {DomainID = 1, WorkflowID = 2, RunID = 3}
// R represents row type in executions table, valid values are:
// R = {Shard = 1, Execution = 2, Transfer = 3, Timer = 4, Replication = 5}
const (
	// Special Domains related constants
	CassandraEmptyDomainID = "10000000-0000-f000-f000-000000000000"
	// Special Run IDs
	CassandraEmptyRunID     = "30000000-0000-f000-f000-000000000000"
	CassandraPermanentRunID = "30000000-0000-f000-f000-000000000001"
	// Row Constants for Shard Row
	CassandraRowTypeShardDomainID   = "10000000-1000-f000-f000-000000000000"
	CassandraRowTypeShardWorkflowID = "20000000-1000-f000-f000-000000000000"
	CassandraRowTypeShardRunID      = "30000000-1000-f000-f000-000000000000"
	// Row Constants for Transfer Task Row
	CassandraRowTypeTransferDomainID   = "10000000-3000-f000-f000-000000000000"
	CassandraRowTypeTransferWorkflowID = "20000000-3000-f000-f000-000000000000"
	CassandraRowTypeTransferRunID      = "30000000-3000-f000-f000-000000000000"
	// Row Constants for Timer Task Row
	CassandraRowTypeTimerDomainID   = "10000000-4000-f000-f000-000000000000"
	CassandraRowTypeTimerWorkflowID = "20000000-4000-f000-f000-000000000000"
	CassandraRowTypeTimerRunID      = "30000000-4000-f000-f000-000000000000"
	// Row Constants for Replication Task Row
	CassandraRowTypeReplicationDomainID   = "10000000-5000-f000-f000-000000000000"
	CassandraRowTypeReplicationWorkflowID = "20000000-5000-f000-f000-000000000000"
	CassandraRowTypeReplicationRunID      = "30000000-5000-f000-f000-000000000000"
	// Row Constants for Replication Task DLQ Row. Source cluster name will be used as WorkflowID.
	CassandraRowTypeDLQDomainID = "10000000-6000-f000-f000-000000000000"
	CassandraRowTypeDLQRunID    = "30000000-6000-f000-f000-000000000000"
	// Special TaskId constants
	CassandraRowTypeExecutionTaskID = int64(-10)
	CassandraRowTypeShardTaskID     = int64(-11)
	CassandraEmptyInitiatedID       = int64(-7)

	CassandraStickyTaskListTTL = int32(24 * time.Hour / time.Second) // if sticky task_list stopped being updated, remove it in one day
)

const (
	// Row types for table tasks
	CassandraRowTypeTask = iota
	CassandraRowTypeTaskList
)

const (
	CassandraTaskListTaskID = -12345
	CassandraInitialRangeID = 1 // Id of the first range of a new task list
)

var (
	CassandraDefaultDateTime            = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	CassandraDefaultVisibilityTimestamp = CassandraUnixNanoToDBTimestamp(CassandraDefaultDateTime.UnixNano())
)

// CassandraDBTimestampToUnixNano converts Milliseconds timestamp to UnixNano
func CassandraDBTimestampToUnixNano(milliseconds int64) int64 {
	return milliseconds * 1000 * 1000 // Milliseconds are 10⁻³, nanoseconds are 10⁻⁹, (-3) - (-9) = 6, so multiply by 10⁶
}

// CassandraUnixNanoToDBTimestamp converts UnixNano to Milliseconds timestamp
func CassandraUnixNanoToDBTimestamp(timestamp int64) int64 {
	return timestamp / (1000 * 1000) // Milliseconds are 10⁻³, nanoseconds are 10⁻⁹, (-9) - (-3) = -6, so divide by 10⁶
}

func CassandraIsTimeoutError(err error) bool {
	if err == gocql.ErrTimeoutNoResponse {
		return true
	}
	if err == gocql.ErrConnectionClosed {
		return true
	}
	_, ok := err.(*gocql.RequestErrWriteTimeout)
	return ok
}

func CassandraIsThrottlingError(err error) bool {
	if req, ok := err.(gocql.RequestError); ok {
		// gocql does not expose the constant errOverloaded = 0x1001
		return req.Code() == 0x1001
	}
	return false
}

const CassandraPersistenceName = "cassandra"


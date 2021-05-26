package cassandra

import (
	"context"
	p "github.com/uber/cadence/common/persistence"
	"time"

	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

const (
	// Row types for table tasks
	rowTypeTask = iota
	rowTypeTaskList
)

const (
	taskListTaskID = -12345
	initialRangeID = 1 // Id of the first range of a new task list
)

const (
	templateTaskListType = `{` +
		`domain_id: ?, ` +
		`name: ?, ` +
		`type: ?, ` +
		`ack_level: ?, ` +
		`kind: ?, ` +
		`last_updated: ? ` +
		`}`

	templateTaskType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`schedule_id: ?,` +
		`created_time: ? ` +
		`}`

	templateCreateTaskQuery = `INSERT INTO tasks (` +
		`domain_id, task_list_name, task_list_type, type, task_id, task) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTaskType + `)`

	templateCreateTaskWithTTLQuery = `INSERT INTO tasks (` +
		`domain_id, task_list_name, task_list_type, type, task_id, task) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTaskType + `) USING TTL ?`

	templateGetTasksQuery = `SELECT task_id, task ` +
		`FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateCompleteTaskQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ?`

	templateCompleteTasksLessThanQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`AND task_list_name = ? ` +
		`AND task_list_type = ? ` +
		`AND type = ? ` +
		`AND task_id <= ? `

	templateGetTaskList = `SELECT ` +
		`range_id, ` +
		`task_list ` +
		`FROM tasks ` +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ?`

	templateInsertTaskListQuery = `INSERT INTO tasks (` +
		`domain_id, ` +
		`task_list_name, ` +
		`task_list_type, ` +
		`type, ` +
		`task_id, ` +
		`range_id, ` +
		`task_list ` +
		`) VALUES (?, ?, ?, ?, ?, ?, ` + templateTaskListType + `) IF NOT EXISTS`

	templateUpdateTaskListQuery = `UPDATE tasks SET ` +
		`range_id = ?, ` +
		`task_list = ` + templateTaskListType + " " +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	templateUpdateTaskListQueryWithTTLPart1 = ` INSERT INTO tasks (` +
		`domain_id, ` +
		`task_list_name, ` +
		`task_list_type, ` +
		`type, ` +
		`task_id ` +
		`) VALUES (?, ?, ?, ?, ?) USING TTL ?`

	templateUpdateTaskListQueryWithTTLPart2 = `UPDATE tasks USING TTL ? SET ` +
		`range_id = ?, ` +
		`task_list = ` + templateTaskListType + " " +
		`WHERE domain_id = ? ` +
		`and task_list_name = ? ` +
		`and task_list_type = ? ` +
		`and type = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	templateDeleteTaskListQuery = `DELETE FROM tasks ` +
		`WHERE domain_id = ? ` +
		`AND task_list_name = ? ` +
		`AND task_list_type = ? ` +
		`AND type = ? ` +
		`AND task_id = ? ` +
		`IF range_id = ?`
)

// SelectTaskList returns a single tasklist row.
// Return IsNotFoundError if the row doesn't exist
func (db *cdb) SelectTaskList(ctx context.Context, filter *nosqlplugin.TaskListFilter) (*nosqlplugin.TaskListRow, error){
	query := db.session.Query(templateGetTaskList,
		filter.DomainID,
		filter.TaskListName,
		filter.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
	).WithContext(ctx)
	var rangeID int64
	var tlDB map[string]interface{}
	err := query.Scan(&rangeID, &tlDB)
	if err != nil{
		return nil, err
	}
	ackLevel := tlDB["ack_level"].(int64)
	taskListKind := tlDB["kind"].(int)
	lastUpdatedTime := tlDB["last_updated"].(time.Time)
	return &nosqlplugin.TaskListRow{
		DomainID: filter.DomainID,
		TaskListName: filter.TaskListName,
		TaskListType: filter.TaskListType,

		TaskListKind: taskListKind,
		LastUpdatedTime: lastUpdatedTime,
		AckLevel: ackLevel,
	}, nil
}
// InsertTaskList insert a single tasklist row
// Return IsConditionFailedError if the row already exists, and also the existing row
func (db *cdb) InsertTaskList(ctx context.Context, row *nosqlplugin.TaskListRow) (*nosqlplugin.TaskListRow, error){
	query := db.session.Query(templateInsertTaskListQuery,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		initialRangeID,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		0,
		row.TaskListKind,
		time.Now(),
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, err
	}
	if !applied {
		rangeID := previous["range_id"]
		taslist := previous["task_list"].(map[string]interface{})
		ackLevel := taslist["ack_level"].(int64)
		taskListKind := taslist["kind"].(int)
		lastUpdatedTime := taslist["last_updated"].(time.Time)

		return nil, &p.ConditionFailedError{
			Msg: fmt.Sprintf("leaseTaskList: taskList:%v, taskListType:%v, haveRangeID:%v, gotRangeID:%v",
				request.TaskList, request.TaskType, rangeID, previousRangeID),
		}
	}
	return nil, nil
}

// UpdateTaskList updates a single tasklist row
// Return IsConditionFailedError if the condition doesn't meet, and also the previous row
func (db *cdb) UpdateTaskList(
	ctx context.Context,
	row *nosqlplugin.TaskListRow,
	previousRangeID int64,
) (*nosqlplugin.TaskListRow, error){

}
// UpdateTaskList updates a single tasklist row, and set an TTL on the record
// Return IsConditionFailedError if the condition doesn't meet, and also the existing row
// Ignore TTL if it's not supported, which becomes exactly the same as UpdateTaskList, but ListTaskList must be
// implemented for TaskListScavenger
func (db *cdb) UpdateTaskListWithTTL(
	ctx context.Context,
	ttlSeconds int64,
	row *nosqlplugin.TaskListRow,
	previousRangeID int64,
) (*nosqlplugin.TaskListRow, error){

}
// ListTaskList returns all tasklists.
// Noop if TTL is already implemented in other methods
func (db *cdb) ListTaskList(ctx context.Context, pageSize int, nextPageToken []byte) (*nosqlplugin.ListTaskListResult, error){

}
// DeleteTaskList deletes a single tasklist row
// Return IsConditionFailedError if the condition doesn't meet, and also the existing row
func (db *cdb) DeleteTaskList(ctx context.Context, filter *nosqlplugin.TaskListFilter) (*nosqlplugin.TaskListRow, error){

}
// InsertTasks inserts a batch of tasks
// Return IsConditionFailedError if the condition doesn't meet, and also the previous tasklist row
func (db *cdb) InsertTasks(
	ctx context.Context,
	tasksToInsert []*nosqlplugin.TaskRowForInsert,
	tasklistCondition *nosqlplugin.TaskListRow,
) (*nosqlplugin.TaskListRow, error){

}
// SelectTasks return tasks that associated to a tasklist
func (db *cdb) SelectTasks(ctx context.Context, filter *nosqlplugin.TasksFilter) ([]*nosqlplugin.TaskRow, error){

}
// DeleteTask delete a single task
func (db *cdb) DeleteTask(ctx context.Context, row *nosqlplugin.TaskRowPK) error{

}

// DeleteTask delete a batch tasks that taskIDs less than or equal to the row
// If TTL is not implemented, then should also return the number of rows deleted, otherwise persistence.UnknownNumRowsAffected
func (db *cdb) RangeDeleteTasks(ctx context.Context, maxTaskID *nosqlplugin.TaskRowPK) (rowsDeleted int, err error){

}

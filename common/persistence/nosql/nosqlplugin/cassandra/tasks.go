package cassandra

import (
	"context"
	"fmt"
	"time"

	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
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
		`and task_id >= ? ` +
		`and task_id < ?`

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
		`AND task_id < ? `

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
		row.LastUpdatedTime,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, err
	}

	return handleTaskListAppliedError(applied, previous)
}

// UpdateTaskList updates a single tasklist row
// Return IsConditionFailedError if the condition doesn't meet, and also the previous row
func (db *cdb) UpdateTaskList(
	ctx context.Context,
	row *nosqlplugin.TaskListRow,
	previousRangeID int64,
) (*nosqlplugin.TaskListRow, error) {
	query := db.session.Query(templateUpdateTaskListQuery,
		row.RangeID,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		row.AckLevel,
		row.TaskListKind,
		row.LastUpdatedTime,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		previousRangeID,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, err
	}

	return handleTaskListAppliedError(applied, previous)
}

func handleTaskListAppliedError(applied bool, previous map[string]interface{}) (*nosqlplugin.TaskListRow, error) {
	if !applied {
		domainID := previous["domain_id"].(gocql.UUID).String()
		taskListName := previous["task_list_name"].(string)
		taskListType := previous["task_list_name"].(int)

		rangeID := previous["range_id"].(int64)
		taslist := previous["task_list"].(map[string]interface{})

		ackLevel := taslist["ack_level"].(int64)
		taskListKind := taslist["kind"].(int)
		lastUpdatedTime := taslist["last_updated"].(time.Time)

		return &nosqlplugin.TaskListRow{
			DomainID:     domainID,
			TaskListName: taskListName,
			TaskListType: taskListType,

			RangeID:         rangeID,
			AckLevel:        ackLevel,
			TaskListKind:    taskListKind,
			LastUpdatedTime: lastUpdatedTime,
		}, errConditionFailed
	}
	return nil, nil
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
) (*nosqlplugin.TaskListRow, error) {
	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	// part 1 is used to set TTL on primary key as UPDATE can't set TTL for primary key
	batch.Query(templateUpdateTaskListQueryWithTTLPart1,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		ttlSeconds,
	)
	// part 2 is for CAS and setting TTL for the rest of the columns
	batch.Query(templateUpdateTaskListQueryWithTTLPart2,
		ttlSeconds,
		row.RangeID,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		row.AckLevel,
		row.TaskListKind,
		time.Now(),
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		previousRangeID,
	)
	previous := make(map[string]interface{})
	applied, _, err := db.session.MapExecuteBatchCAS(batch, previous)
	if err != nil {
		return nil, err
	}
	return handleTaskListAppliedError(applied, previous)
}

// ListTaskList returns all tasklists.
// Noop if TTL is already implemented in other methods
func (db *cdb) ListTaskList(ctx context.Context, pageSize int, nextPageToken []byte) (*nosqlplugin.ListTaskListResult, error) {
	return nil, &types.InternalServiceError{
		Message: fmt.Sprintf("unsupported operation"),
	}
}

// DeleteTaskList deletes a single tasklist row
// Return IsConditionFailedError if the condition doesn't meet, and also the existing row
func (db *cdb) DeleteTaskList(ctx context.Context, filter *nosqlplugin.TaskListFilter, previousRangeID int64) (*nosqlplugin.TaskListRow, error) {
	query := db.session.Query(templateDeleteTaskListQuery,
		filter.DomainID,
		filter.TaskListName,
		filter.TaskListType,
		rowTypeTaskList,
		taskListTaskID,
		previousRangeID,
	).WithContext(ctx)
	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, err
	}
	return handleTaskListAppliedError(applied, previous)
}

// InsertTasks inserts a batch of tasks
// Return IsConditionFailedError if the condition doesn't meet, and also the previous tasklist row
func (db *cdb) InsertTasks(
	ctx context.Context,
	tasksToInsert []*nosqlplugin.TaskRowForInsert,
	tasklistCondition *nosqlplugin.TaskListRow,
) (*nosqlplugin.TaskListRow, error) {
	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	domainID := tasklistCondition.DomainID
	taskListName := tasklistCondition.TaskListName
	taskListType := tasklistCondition.TaskListType
	taskListKind := tasklistCondition.TaskListKind
	ackLevel := tasklistCondition.AckLevel

	for _, task := range tasksToInsert {
		scheduleID := task.ScheduledID
		ttl := int64(task.TTLSeconds)
		if ttl <= 0 {
			batch.Query(templateCreateTaskQuery,
				domainID,
				taskListName,
				taskListType,
				rowTypeTask,
				task.TaskID,
				domainID,
				task.WorkflowID,
				task.RunID,
				scheduleID,
				task.CreatedTime)
		} else {
			if ttl > maxCassandraTTL {
				ttl = maxCassandraTTL
			}
			batch.Query(templateCreateTaskWithTTLQuery,
				domainID,
				taskListName,
				taskListType,
				rowTypeTask,
				task.TaskID,
				domainID,
				task.WorkflowID,
				task.RunID,
				scheduleID,
				tasklistCondition.LastUpdatedTime,
				ttl)
		}
	}

	// The following query is used to ensure that range_id didn't change
	batch.Query(templateUpdateTaskListQuery,
		tasklistCondition.RangeID,
		domainID,
		taskListName,
		taskListType,
		ackLevel,
		taskListKind,
		time.Now(),
		domainID,
		taskListName,
		taskListType,
		rowTypeTaskList,
		taskListTaskID,
		tasklistCondition.RangeID,
	)

	previous := make(map[string]interface{})
	applied, _, err := db.session.MapExecuteBatchCAS(batch, previous)
	if err != nil {
		return nil, err
	}
	return handleTaskListAppliedError(applied, previous)
}

// SelectTasks return tasks that associated to a tasklist
func (db *cdb) SelectTasks(ctx context.Context, filter *nosqlplugin.TasksFilter) ([]*nosqlplugin.TaskRow, error) {
	// Reading tasklist tasks need to be quorum level consistent, otherwise we could loose task
	query := db.session.Query(templateGetTasksQuery,
		filter.DomainID,
		filter.TaskListName,
		filter.TaskListType,
		rowTypeTask,
		filter.MinTaskID,
		filter.MaxTaskID,
	).PageSize(filter.BatchSize).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, fmt.Errorf("GetTasks operation failed.  Not able to create query iterator.")
	}

	var response []*nosqlplugin.TaskRow
	task := make(map[string]interface{})
PopulateTasks:
	for iter.MapScan(task) {
		taskID, ok := task["task_id"]
		if !ok { // no tasks, but static column record returned
			continue
		}
		t := createTaskInfo(task["task"].(map[string]interface{}))
		t.TaskID = taskID.(int64)
		response = append(response, t)
		if len(response) == filter.BatchSize {
			break PopulateTasks
		}
		task = make(map[string]interface{}) // Reinitialize map as initialized fails on unmarshalling
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return response, nil
}

func createTaskInfo(
	result map[string]interface{},
) *nosqlplugin.TaskRow {

	info := &nosqlplugin.TaskRow{}
	for k, v := range result {
		switch k {
		case "domain_id":
			info.DomainID = v.(gocql.UUID).String()
		case "workflow_id":
			info.WorkflowID = v.(string)
		case "run_id":
			info.RunID = v.(gocql.UUID).String()
		case "schedule_id":
			info.ScheduledID = v.(int64)
		case "created_time":
			info.CreatedTime = v.(time.Time)
		}
	}

	return info
}

// DeleteTask delete a single task
func (db *cdb) DeleteTask(ctx context.Context, row *nosqlplugin.TaskRowPK) error {
	query := db.session.Query(templateCompleteTaskQuery,
		row.DomainID,
		row.TaskListName,
		row.TaskListType,
		rowTypeTask,
		row.TaskID,
	).WithContext(ctx)

	return query.Exec()
}

// DeleteTask delete a batch tasks that taskIDs less than the row
// If TTL is not implemented, then should also return the number of rows deleted, otherwise persistence.UnknownNumRowsAffected
// NOTE: This API ignores the `Limit` request parameter i.e. either all tasks leq the task_id will be deleted or an error will
// be returned to the caller, because rowsDeleted is not supported by Cassandra
func (db *cdb) RangeDeleteTasks(ctx context.Context, maxTaskID *nosqlplugin.TaskRowPK, _ int) (rowsDeleted int, err error) {
	query := db.session.Query(templateCompleteTasksLessThanQuery,
		maxTaskID.DomainID,
		maxTaskID.TaskListName,
		maxTaskID.TaskListType,
		rowTypeTask,
		maxTaskID.TaskID,
	).WithContext(ctx)
	err = query.Exec()
	return p.UnknownNumRowsAffected, err
}

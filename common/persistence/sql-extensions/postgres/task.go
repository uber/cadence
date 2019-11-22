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

const (
	taskListCreatePart = `INTO task_lists(shard_id, domain_id, name, task_type, range_id, data, data_encoding) ` +
		`VALUES (:shard_id, :domain_id, :name, :task_type, :range_id, :data, :data_encoding)`

	// (default range ID: initialRangeID == 1)
	createTaskListQry = `INSERT ` + taskListCreatePart

	replaceTaskListQry = `REPLACE ` + taskListCreatePart

	updateTaskListQry = `UPDATE task_lists SET
range_id = :range_id,
data = :data,
data_encoding = :data_encoding
WHERE
shard_id = :shard_id AND
domain_id = :domain_id AND
name = :name AND
task_type = :task_type
`

	listTaskListQry = `SELECT domain_id, range_id, name, task_type, data, data_encoding ` +
		`FROM task_lists ` +
		`WHERE shard_id = $1 AND domain_id > $2 AND name > $3 AND task_type > $4 ORDER BY domain_id,name,task_type LIMIT $5`

	getTaskListQry = `SELECT domain_id, range_id, name, task_type, data, data_encoding ` +
		`FROM task_lists ` +
		`WHERE shard_id = $1 AND domain_id = $2 AND name = $3 AND task_type = $4`

	deleteTaskListQry = `DELETE FROM task_lists WHERE shard_id=$1 AND domain_id=$2 AND name=$3 AND task_type=$4 AND range_id=$5`

	lockTaskListQry = `SELECT range_id FROM task_lists ` +
		`WHERE shard_id = $1 AND domain_id = $2 AND name = $3 AND task_type = $4 FOR UPDATE`

	getTaskMinMaxQry = `SELECT task_id, data, data_encoding ` +
		`FROM tasks ` +
		`WHERE domain_id = $1 AND task_list_name = $2 AND task_type = $3 AND task_id > $4 AND task_id <= $5 ` +
		` ORDER BY task_id LIMIT $6`

	getTaskMinQry = `SELECT task_id, data, data_encoding ` +
		`FROM tasks ` +
		`WHERE domain_id = $1 AND task_list_name = $2 AND task_type = $3 AND task_id > $4 ORDER BY task_id LIMIT $5`

	createTaskQry = `INSERT INTO ` +
		`tasks(domain_id, task_list_name, task_type, task_id, data, data_encoding) ` +
		`VALUES(:domain_id, :task_list_name, :task_type, :task_id, :data, :data_encoding)`

	deleteTaskQry = `DELETE FROM tasks ` +
		`WHERE domain_id = $1 AND task_list_name = $2 AND task_type = $3 AND task_id = $4`

	rangeDeleteTaskQry = `DELETE FROM tasks ` +
		`WHERE domain_id = $1 AND task_list_name = $2 AND task_type = $3 AND task_id <= $4 ` +
		`ORDER BY domain_id,task_list_name,task_type,task_id LIMIT $5`
)

func (d *driver) CreateTaskListQuery() string {
	return createTaskListQry
}
func (d *driver) ReplaceTaskListQuery() string {
	return replaceTaskListQry
}
func (d *driver) UpdateTaskListQuery() string {
	return updateTaskListQry
}
func (d *driver) ListTaskListQuery() string {
	return listTaskListQry
}
func (d *driver) GetTaskListQuery() string {
	return getTaskListQry
}
func (d *driver) DeleteTaskListQuery() string {
	return deleteTaskListQry
}
func (d *driver) LockTaskListQuery() string {
	return lockTaskListQry
}
func (d *driver) GetTaskMinMaxQuery() string {
	return getTaskMinMaxQry
}
func (d *driver) GetTaskMinQuery() string {
	return getTaskMinQry
}
func (d *driver) CreateTaskQuery() string {
	return createTaskQry
}
func (d *driver) DeleteTaskQuery() string {
	return deleteTaskQry
}
func (d *driver) RangeDeleteTaskQuery() string {
	return rangeDeleteTaskQry
}

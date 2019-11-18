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
	"fmt"
	"strings"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

func stringMap(a []string, f func(string) string) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = f(v)
	}
	return b
}

func makeDeleteMapQry(driver Driver, tableName string) string {
	return fmt.Sprintf(driver.DeleteMapQueryTemplate(), tableName)
}

func makeSetKeyInMapQry(driver Driver, tableName string, nonPrimaryKeyColumns []string, mapKeyName string) string {
	return fmt.Sprintf(driver.SetKeyInMapQueryTemplate(),
		tableName,
		strings.Join(nonPrimaryKeyColumns, ","),
		strings.Join(stringMap(nonPrimaryKeyColumns, func(x string) string {
			return ":" + x
		}), ","),
		mapKeyName)
}

func makeDeleteKeyInMapQry(driver Driver, tableName string, mapKeyName string) string {
	return fmt.Sprintf(driver.DeleteKeyInMapQueryTemplate(),
		tableName,
		mapKeyName)
}

func makeGetMapQryTemplate(driver Driver, tableName string, nonPrimaryKeyColumns []string, mapKeyName string) string {
	return fmt.Sprintf(driver.GetMapQueryTemplate(),
		tableName,
		mapKeyName,
		strings.Join(nonPrimaryKeyColumns, ","))
}

var (
	// Omit shard_id, run_id, domain_id, workflow_id, schedule_id since they're in the primary key
	activityInfoColumns = []string{
		"data",
		"data_encoding",
		"last_heartbeat_details",
		"last_heartbeat_updated_time",
	}
	activityInfoTableName = "activity_info_maps"
	activityInfoKey       = "schedule_id"
)

// ReplaceIntoActivityInfoMaps replaces one or more rows in activity_info_maps table
func (mdb *DB) ReplaceIntoActivityInfoMaps(rows []sqldb.ActivityInfoMapsRow) (sql.Result, error) {
	for i := range rows {
		rows[i].LastHeartbeatUpdatedTime = mdb.converter.ToMySQLDateTime(rows[i].LastHeartbeatUpdatedTime)
	}
	return mdb.conn.NamedExec(makeSetKeyInMapQry(mdb.driver, activityInfoTableName, activityInfoColumns, activityInfoKey), rows)
}

// SelectFromActivityInfoMaps reads one or more rows from activity_info_maps table
func (mdb *DB) SelectFromActivityInfoMaps(filter *sqldb.ActivityInfoMapsFilter) ([]sqldb.ActivityInfoMapsRow, error) {
	var rows []sqldb.ActivityInfoMapsRow
	err := mdb.conn.Select(&rows, makeGetMapQryTemplate(mdb.driver, activityInfoTableName, activityInfoColumns, activityInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
		rows[i].LastHeartbeatUpdatedTime = mdb.converter.FromMySQLDateTime(rows[i].LastHeartbeatUpdatedTime)
	}
	return rows, err
}

// DeleteFromActivityInfoMaps deletes one or more rows from activity_info_maps table
func (mdb *DB) DeleteFromActivityInfoMaps(filter *sqldb.ActivityInfoMapsFilter) (sql.Result, error) {
	if filter.ScheduleID != nil {
		return mdb.conn.Exec(makeDeleteKeyInMapQry(mdb.driver, activityInfoTableName, activityInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.ScheduleID)
	}
	return mdb.conn.Exec(makeDeleteMapQry(mdb.driver, activityInfoTableName), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

var (
	timerInfoColumns = []string{
		"data",
		"data_encoding",
	}
	timerInfoTableName = "timer_info_maps"
	timerInfoKey       = "timer_id"
)

// ReplaceIntoTimerInfoMaps replaces one or more rows in timer_info_maps table
func (mdb *DB) ReplaceIntoTimerInfoMaps(rows []sqldb.TimerInfoMapsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(makeSetKeyInMapQry(mdb.driver, timerInfoTableName, timerInfoColumns, timerInfoKey), rows)
}

// SelectFromTimerInfoMaps reads one or more rows from timer_info_maps table
func (mdb *DB) SelectFromTimerInfoMaps(filter *sqldb.TimerInfoMapsFilter) ([]sqldb.TimerInfoMapsRow, error) {
	var rows []sqldb.TimerInfoMapsRow
	err := mdb.conn.Select(&rows, makeGetMapQryTemplate(mdb.driver, timerInfoTableName, timerInfoColumns, timerInfoKey),
							filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
	}
	return rows, err
}

// DeleteFromTimerInfoMaps deletes one or more rows from timer_info_maps table
func (mdb *DB) DeleteFromTimerInfoMaps(filter *sqldb.TimerInfoMapsFilter) (sql.Result, error) {
	if filter.TimerID != nil {
		return mdb.conn.Exec(makeDeleteKeyInMapQry(mdb.driver, timerInfoTableName, timerInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.TimerID)
	}
	return mdb.conn.Exec(makeDeleteMapQry(mdb.driver, timerInfoTableName), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

var (
	childExecutionInfoColumns = []string{
		"data",
		"data_encoding",
	}
	childExecutionInfoTableName = "child_execution_info_maps"
	childExecutionInfoKey       = "initiated_id"
)

// ReplaceIntoChildExecutionInfoMaps replaces one or more rows in child_execution_info_maps table
func (mdb *DB) ReplaceIntoChildExecutionInfoMaps(rows []sqldb.ChildExecutionInfoMapsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(makeSetKeyInMapQry(mdb.driver, childExecutionInfoTableName, childExecutionInfoColumns, childExecutionInfoKey), rows)
}

// SelectFromChildExecutionInfoMaps reads one or more rows from child_execution_info_maps table
func (mdb *DB) SelectFromChildExecutionInfoMaps(filter *sqldb.ChildExecutionInfoMapsFilter) ([]sqldb.ChildExecutionInfoMapsRow, error) {
	var rows []sqldb.ChildExecutionInfoMapsRow
	err := mdb.conn.Select(&rows, makeGetMapQryTemplate(mdb.driver, childExecutionInfoTableName, childExecutionInfoColumns, childExecutionInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
	}
	return rows, err
}

// DeleteFromChildExecutionInfoMaps deletes one or more rows from child_execution_info_maps table
func (mdb *DB) DeleteFromChildExecutionInfoMaps(filter *sqldb.ChildExecutionInfoMapsFilter) (sql.Result, error) {
	if filter.InitiatedID != nil {
		return mdb.conn.Exec(makeDeleteKeyInMapQry(mdb.driver, childExecutionInfoTableName, childExecutionInfoKey),
			filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.InitiatedID)
	}
	return mdb.conn.Exec(makeDeleteMapQry(mdb.driver, childExecutionInfoTableName), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

var (
	requestCancelInfoColumns = []string{
		"data",
		"data_encoding",
	}
	requestCancelInfoTableName = "request_cancel_info_maps"
	requestCancelInfoKey       = "initiated_id"
)

// ReplaceIntoRequestCancelInfoMaps replaces one or more rows in request_cancel_info_maps table
func (mdb *DB) ReplaceIntoRequestCancelInfoMaps(rows []sqldb.RequestCancelInfoMapsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(makeSetKeyInMapQry(mdb.driver, requestCancelInfoTableName, requestCancelInfoColumns, requestCancelInfoKey), rows)
}

// SelectFromRequestCancelInfoMaps reads one or more rows from request_cancel_info_maps table
func (mdb *DB) SelectFromRequestCancelInfoMaps(filter *sqldb.RequestCancelInfoMapsFilter) ([]sqldb.RequestCancelInfoMapsRow, error) {
	var rows []sqldb.RequestCancelInfoMapsRow
	err := mdb.conn.Select(&rows, makeGetMapQryTemplate(mdb.driver, requestCancelInfoTableName, requestCancelInfoColumns, requestCancelInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
	}
	return rows, err
}

// DeleteFromRequestCancelInfoMaps deletes one or more rows from request_cancel_info_maps table
func (mdb *DB) DeleteFromRequestCancelInfoMaps(filter *sqldb.RequestCancelInfoMapsFilter) (sql.Result, error) {
	if filter.InitiatedID != nil {
		return mdb.conn.Exec(makeDeleteKeyInMapQry(mdb.driver, requestCancelInfoTableName, requestCancelInfoKey), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.InitiatedID)
	}
	return mdb.conn.Exec(makeDeleteMapQry(mdb.driver, requestCancelInfoTableName), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

var (
	signalInfoColumns = []string{
		"data",
		"data_encoding",
	}
	signalInfoTableName = "signal_info_maps"
	signalInfoKey       = "initiated_id"
)

// ReplaceIntoSignalInfoMaps replaces one or more rows in signal_info_maps table
func (mdb *DB) ReplaceIntoSignalInfoMaps(rows []sqldb.SignalInfoMapsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(makeSetKeyInMapQry(mdb.driver, signalInfoTableName, signalInfoColumns, signalInfoKey), rows)
}

// SelectFromSignalInfoMaps reads one or more rows from signal_info_maps table
func (mdb *DB) SelectFromSignalInfoMaps(filter *sqldb.SignalInfoMapsFilter) ([]sqldb.SignalInfoMapsRow, error) {
	var rows []sqldb.SignalInfoMapsRow
	err := mdb.conn.Select(&rows, makeGetMapQryTemplate(mdb.driver, signalInfoTableName, signalInfoColumns, signalInfoKey),
		filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
	}
	return rows, err
}

// DeleteFromSignalInfoMaps deletes one or more rows from signal_info_maps table
func (mdb *DB) DeleteFromSignalInfoMaps(filter *sqldb.SignalInfoMapsFilter) (sql.Result, error) {
	if filter.InitiatedID != nil {
		return mdb.conn.Exec(makeDeleteKeyInMapQry(mdb.driver, signalInfoTableName, signalInfoKey),
			filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.InitiatedID)
	}
	return mdb.conn.Exec(makeDeleteMapQry(mdb.driver, signalInfoTableName), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// InsertIntoSignalsRequestedSets inserts one or more rows into signals_requested_sets table
func (mdb *DB) InsertIntoSignalsRequestedSets(rows []sqldb.SignalsRequestedSetsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateSignalsRequestedSetQuery(), rows)
}

// SelectFromSignalsRequestedSets reads one or more rows from signals_requested_sets table
func (mdb *DB) SelectFromSignalsRequestedSets(filter *sqldb.SignalsRequestedSetsFilter) ([]sqldb.SignalsRequestedSetsRow, error) {
	var rows []sqldb.SignalsRequestedSetsRow
	err := mdb.conn.Select(&rows, mdb.driver.GetSignalsRequestedSetQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].ShardID = int64(filter.ShardID)
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
	}
	return rows, err
}

// DeleteFromSignalsRequestedSets deletes one or more rows from signals_requested_sets table
func (mdb *DB) DeleteFromSignalsRequestedSets(filter *sqldb.SignalsRequestedSetsFilter) (sql.Result, error) {
	if filter.SignalID != nil {
		return mdb.conn.Exec(mdb.driver.DeleteSignalsRequestedSetQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID, *filter.SignalID)
	}
	return mdb.conn.Exec(mdb.driver.DeleteAllSignalsRequestedSetQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

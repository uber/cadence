package cli

import (
	"github.com/uber/cadence/.gen/go/config"
	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common/codec"
)

var decodingTypes = map[string]codec.ThriftObject{
	"shared.History": &shared.History{},
	"shared.HistoryEvent": &shared.HistoryEvent{},
	"shared.Memo": &shared.Memo{},
	"shared.ResetPoints": &shared.ResetPoints{},
	"shared.BadBinaries": &shared.BadBinaries{},
	"shared.VersionHistories": &shared.VersionHistories{},
	"replicator.FailoverMarkers": &replicator.FailoverMarkers{},
	"history.ProcessingQueueStates": &history.ProcessingQueueStates{},
	"config.DynamicConfigBlob": &config.DynamicConfigBlob{},
	"sqlblobs.ShardInfo": &sqlblobs.ShardInfo{},
	"sqlblobs.DomainInfo": &sqlblobs.DomainInfo{},
	"sqlblobs.HistoryTreeInfo": &sqlblobs.HistoryTreeInfo{},
	"sqlblobs.WorkflowExecutionInfo": &sqlblobs.WorkflowExecutionInfo{},
	"sqlblobs.ActivityInfo": &sqlblobs.ActivityInfo{},
	"sqlblobs.ChildExecutionInfo": &sqlblobs.ChildExecutionInfo{},
	"sqlblobs.SignalInfo": &sqlblobs.SignalInfo{},
	"sqlblobs.RequestCancelInfo": &sqlblobs.RequestCancelInfo{},
	"sqlblobs.TimerInfo": &sqlblobs.TimerInfo{},
	"sqlblobs.TaskInfo": &sqlblobs.TaskInfo{},
	"sqlblobs.TaskListInfo": &sqlblobs.TaskListInfo{},
	"sqlblobs.TransferTaskInfo": &sqlblobs.TransferTaskInfo{},
	"sqlblobs.TimerTaskInfo": &sqlblobs.TimerTaskInfo{},
	"sqlblobs.ReplicationTaskInfo": &sqlblobs.ReplicationTaskInfo{},
}
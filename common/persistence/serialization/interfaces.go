package serialization

import (
	"github.com/uber/cadence/.gen/go/sqlblobs"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// Encoder is used to serialize to persistence data blobs
	Encoder interface {
		ShardInfoToBlob(*sqlblobs.ShardInfo) (p.DataBlob, error)
		DomainInfoToBlob(*sqlblobs.DomainInfo) (p.DataBlob, error)
		HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) (p.DataBlob, error)
		WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) (p.DataBlob, error)
		ActivityInfoToBlob(*sqlblobs.ActivityInfo) (p.DataBlob, error)
		ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) (p.DataBlob, error)
		SignalInfoToBlob(*sqlblobs.SignalInfo) (p.DataBlob, error)
		RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) (p.DataBlob, error)
		TimerInfoToBlob(*sqlblobs.TimerInfo) (p.DataBlob, error)
		TaskInfoToBlob(*sqlblobs.TaskInfo) (p.DataBlob, error)
		TaskListInfoToBlob(*sqlblobs.TaskListInfo) (p.DataBlob, error)
		TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) (p.DataBlob, error)
		TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) (p.DataBlob, error)
		ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) (p.DataBlob, error)
	}

	// Decoder is used to deserialize from persistence structs
	Decoder interface {
		ShardInfoFromBlob([]byte, string) (*sqlblobs.ShardInfo, error)
		DomainInfoFromBlob([]byte, string) (*sqlblobs.DomainInfo, error)
		HistoryTreeInfoFromBlob([]byte, string) (*sqlblobs.HistoryTreeInfo, error)
		WorkflowExecutionInfoFromBlob([]byte, string) (*sqlblobs.WorkflowExecutionInfo, error)
		ActivityInfoFromBlob([]byte, string) (*sqlblobs.ActivityInfo, error)
		ChildExecutionInfoFromBlob([]byte, string) (*sqlblobs.ChildExecutionInfo, error)
		SignalInfoFromBlob([]byte, string) (*sqlblobs.SignalInfo, error)
		RequestCancelInfoFromBlob([]byte, string) (*sqlblobs.RequestCancelInfo, error)
		TimerInfoFromBlob([]byte, string) (*sqlblobs.TimerInfo, error)
		TaskInfoFromBlob([]byte, string) (*sqlblobs.TaskInfo, error)
		TaskListInfoFromBlob(b []byte, proto string) (*sqlblobs.TaskListInfo, error)
		TransferTaskInfoFromBlob([]byte, string) (*sqlblobs.TransferTaskInfo, error)
		TimerTaskInfoFromBlob([]byte, string) (*sqlblobs.TimerTaskInfo, error)
		ReplicationTaskInfoFromBlob([]byte, string) (*sqlblobs.ReplicationTaskInfo, error)
	}
)
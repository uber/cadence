package serialization

import (
	"github.com/uber/cadence/.gen/go/sqlblobs"
	p "github.com/uber/cadence/common/persistence"
)

type thriftEncoder struct {}

// NewThriftEncoder returns a new thrift encoder
func NewThriftEncoder() Encoder {
	return &thriftEncoder{}
}

func (e *thriftEncoder) ShardInfoToBlob(info *sqlblobs.ShardInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) DomainInfoToBlob(info *sqlblobs.DomainInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) HistoryTreeInfoToBlob(info *sqlblobs.HistoryTreeInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) WorkflowExecutionInfoToBlob(info *sqlblobs.WorkflowExecutionInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ActivityInfoToBlob(info *sqlblobs.ActivityInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ChildExecutionInfoToBlob(info *sqlblobs.ChildExecutionInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) SignalInfoToBlob(info *sqlblobs.SignalInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) RequestCancelInfoToBlob(info *sqlblobs.RequestCancelInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TimerInfoToBlob(info *sqlblobs.TimerInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TaskInfoToBlob(info *sqlblobs.TaskInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TaskListInfoToBlob(info *sqlblobs.TaskListInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TransferTaskInfoToBlob(info *sqlblobs.TransferTaskInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TimerTaskInfoToBlob(info *sqlblobs.TimerTaskInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ReplicationTaskInfoToBlob(info *sqlblobs.ReplicationTaskInfo) (p.DataBlob, error) {
	return thriftRWEncode(info)
}

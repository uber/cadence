package serialization

import (
	"github.com/uber/cadence/.gen/go/sqlblobs"
	p "github.com/uber/cadence/common/persistence"
)

type protoEncoder struct {}

// NewProtoEncoder returns a new proto encoder
func NewProtoEncoder() Encoder {
	return &protoEncoder{}
}

func (e *protoEncoder) ShardInfoToBlob(*sqlblobs.ShardInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) DomainInfoToBlob(*sqlblobs.DomainInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) ActivityInfoToBlob(*sqlblobs.ActivityInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) SignalInfoToBlob(*sqlblobs.SignalInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) TimerInfoToBlob(*sqlblobs.TimerInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) TaskInfoToBlob(*sqlblobs.TaskInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) TaskListInfoToBlob(*sqlblobs.TaskListInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) (p.DataBlob, error) {
	panic("not implemented")
}

func (e *protoEncoder) ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) (p.DataBlob, error) {
	panic("not implemented")
}

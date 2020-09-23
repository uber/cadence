package serialization

import "github.com/uber/cadence/.gen/go/sqlblobs"

type (
	thriftDecoder struct {}
)

// NewThriftDecoder creates a decoder which knows how to decode thrift serialized blobs
func NewThriftDecoder() Decoder {
	return &thriftDecoder{}
}

func (d *thriftDecoder) ShardInfoFromBlob([]byte) (*sqlblobs.ShardInfo, error) {

}

func (d *thriftDecoder) DomainInfoFromBlob([]byte) (*sqlblobs.DomainInfo, error) {

}

func (d *thriftDecoder) HistoryTreeInfoFromBlob([]byte) (*sqlblobs.HistoryTreeInfo, error) {

}

func (d *thriftDecoder) WorkflowExecutionInfoFromBlob([]byte) (*sqlblobs.WorkflowExecutionInfo, error) {

}

func (d *thriftDecoder) ActivityInfoFromBlob([]byte) (*sqlblobs.ActivityInfo, error) {

}

func (d *thriftDecoder) ChildExecutionInfoFromBlob([]byte) (*sqlblobs.ChildExecutionInfo, error) {

}

func (d *thriftDecoder) SignalInfoFromBlob([]byte) (*sqlblobs.SignalInfo, error) {

}

func (d *thriftDecoder) RequestCancelInfoFromBlob([]byte) (*sqlblobs.RequestCancelInfo, error) {

}

func (d *thriftDecoder) TimerInfoFromBlob([]byte) (*sqlblobs.TimerInfo, error) {

}

func (d *thriftDecoder) TaskInfoFromBlob([]byte) (*sqlblobs.TaskInfo, error) {

}

func (d *thriftDecoder) TaskListInfoFromBlob([]byte) (*sqlblobs.TaskListInfo, error) {

}

func (d *thriftDecoder) TransferTaskInfoFromBlob([]byte) (*sqlblobs.TransferTaskInfo, error) {

}

func (d *thriftDecoder) TimerTaskInfoFromBlob([]byte) (*sqlblobs.TimerTaskInfo, error) {

}

func (d *thriftDecoder) ReplicationTaskInfoFromBlob([]byte) (*sqlblobs.ReplicationTaskInfo, error) {

}
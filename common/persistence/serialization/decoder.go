package serialization

import (
	"fmt"
	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
)

type (
	decoder struct {}
)

// NewDecoder returns a new decoder
func NewDecoder() Decoder {
	return &decoder{}
}

func (d *decoder) ShardInfoFromBlob(data []byte, encoding string) (*sqlblobs.ShardInfo, error) {
	result := &sqlblobs.ShardInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) DomainInfoFromBlob(data []byte, encoding string) (*sqlblobs.DomainInfo, error) {
	result := &sqlblobs.DomainInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) HistoryTreeInfoFromBlob(data []byte, encoding string) (*sqlblobs.HistoryTreeInfo, error)  {
	result := &sqlblobs.HistoryTreeInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) workflowExecutionInfoFromBlob(data []byte, encoding string) (*sqlblobs.WorkflowExecutionInfo, error) {
	result := &sqlblobs.WorkflowExecutionInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) ActivityInfoFromBlob(data []byte, encoding string) (*sqlblobs.ActivityInfo, error) {
	result := &sqlblobs.ActivityInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) ChildExecutionInfoFromBlob(data []byte, encoding string) (*sqlblobs.ChildExecutionInfo, error) {
	result := &sqlblobs.ChildExecutionInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) SignalInfoFromBlob(data []byte, encoding string) (*sqlblobs.SignalInfo, error) {
	result := &sqlblobs.SignalInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) RequestCancelInfoFromBlob(data []byte, encoding string) (*sqlblobs.RequestCancelInfo, error) {
	result := &sqlblobs.RequestCancelInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) TimerInfoFromBlob(data []byte, encoding string) (*sqlblobs.TimerInfo, error) {
	result := &sqlblobs.TimerInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) TaskInfoFromBlob(data []byte, encoding string) (*sqlblobs.TaskInfo, error) {
	result := &sqlblobs.TaskInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) TaskListInfoFromBlob(data []byte, encoding string) (*sqlblobs.TaskListInfo, error) {
	result := &sqlblobs.TaskListInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) TransferTaskInfoFromBlob(data []byte, encoding string) (*sqlblobs.TransferTaskInfo, error) {
	result := &sqlblobs.TransferTaskInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) TimerTaskInfoFromBlob(data []byte, encoding string) (*sqlblobs.TimerTaskInfo, error) {
	result := &sqlblobs.TimerTaskInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}

func (d *decoder) ReplicationTaskInfoFromBlob(data []byte, encoding string) (*sqlblobs.ReplicationTaskInfo, error) {
	result := &sqlblobs.ReplicationTaskInfo{}
	if err := validateEncoding(encoding); err != nil {
		return result, err
	}
	switch common.EncodingType(encoding) {
	case common.EncodingTypeThriftRW:
		return result, thriftRWDecode(data, result)
	case common.EncodingTypeProto:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", encoding))
	}
}
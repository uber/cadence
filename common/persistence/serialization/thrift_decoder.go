// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package serialization

import (
	"bytes"

	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"

	"github.com/uber/cadence/.gen/go/sqlblobs"
)

type (
	thriftDecoder struct{}
)

func newThriftDecoder() decoder {
	return &thriftDecoder{}
}

func (d *thriftDecoder) shardInfoFromBlob(data []byte) (*sqlblobs.ShardInfo, error) {
	result := &sqlblobs.ShardInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) domainInfoFromBlob(data []byte) (*sqlblobs.DomainInfo, error) {
	result := &sqlblobs.DomainInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) historyTreeInfoFromBlob(data []byte) (*sqlblobs.HistoryTreeInfo, error) {
	result := &sqlblobs.HistoryTreeInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) workflowExecutionInfoFromBlob(data []byte) (*sqlblobs.WorkflowExecutionInfo, error) {
	result := &sqlblobs.WorkflowExecutionInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) activityInfoFromBlob(data []byte) (*sqlblobs.ActivityInfo, error) {
	result := &sqlblobs.ActivityInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) childExecutionInfoFromBlob(data []byte) (*sqlblobs.ChildExecutionInfo, error) {
	result := &sqlblobs.ChildExecutionInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) signalInfoFromBlob(data []byte) (*sqlblobs.SignalInfo, error) {
	result := &sqlblobs.SignalInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) requestCancelInfoFromBlob(data []byte) (*sqlblobs.RequestCancelInfo, error) {
	result := &sqlblobs.RequestCancelInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) timerInfoFromBlob(data []byte) (*sqlblobs.TimerInfo, error) {
	result := &sqlblobs.TimerInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) taskInfoFromBlob(data []byte) (*sqlblobs.TaskInfo, error) {
	result := &sqlblobs.TaskInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) taskListInfoFromBlob(data []byte) (*sqlblobs.TaskListInfo, error) {
	result := &sqlblobs.TaskListInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) transferTaskInfoFromBlob(data []byte) (*sqlblobs.TransferTaskInfo, error) {
	result := &sqlblobs.TransferTaskInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) timerTaskInfoFromBlob(data []byte) (*sqlblobs.TimerTaskInfo, error) {
	result := &sqlblobs.TimerTaskInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *thriftDecoder) replicationTaskInfoFromBlob(data []byte) (*sqlblobs.ReplicationTaskInfo, error) {
	result := &sqlblobs.ReplicationTaskInfo{}
	if err := thriftRWDecode(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func thriftRWDecode(b []byte, result thriftRWType) error {
	value, err := protocol.Binary.Decode(bytes.NewReader(b), wire.TStruct)
	if err != nil {
		return err
	}
	return result.FromWire(value)
}

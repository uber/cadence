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

	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
)

type thriftEncoder struct{}

func newThriftEncoder() encoder {
	return &thriftEncoder{}
}

func (e *thriftEncoder) shardInfoToBlob(info *sqlblobs.ShardInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) domainInfoToBlob(info *sqlblobs.DomainInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) historyTreeInfoToBlob(info *sqlblobs.HistoryTreeInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) workflowExecutionInfoToBlob(info *sqlblobs.WorkflowExecutionInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) activityInfoToBlob(info *sqlblobs.ActivityInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) childExecutionInfoToBlob(info *sqlblobs.ChildExecutionInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) signalInfoToBlob(info *sqlblobs.SignalInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) requestCancelInfoToBlob(info *sqlblobs.RequestCancelInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) timerInfoToBlob(info *sqlblobs.TimerInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) taskInfoToBlob(info *sqlblobs.TaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) taskListInfoToBlob(info *sqlblobs.TaskListInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) transferTaskInfoToBlob(info *sqlblobs.TransferTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) timerTaskInfoToBlob(info *sqlblobs.TimerTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) replicationTaskInfoToBlob(info *sqlblobs.ReplicationTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) encodingType() common.EncodingType {
	return common.EncodingTypeThriftRW
}

func thriftRWEncode(t thriftRWType) ([]byte, error) {
	value, err := t.ToWire()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := protocol.Binary.Encode(value, &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

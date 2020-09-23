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
	"github.com/uber/cadence/.gen/go/sqlblobs"
	"go.uber.org/thriftrw/protocol"
)

type thriftEncoder struct{}

// NewThriftEncoder returns a new thrift encoder
func NewThriftEncoder() Encoder {
	return &thriftEncoder{}
}

func (e *thriftEncoder) ShardInfoToBlob(info *sqlblobs.ShardInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) DomainInfoToBlob(info *sqlblobs.DomainInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) HistoryTreeInfoToBlob(info *sqlblobs.HistoryTreeInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) WorkflowExecutionInfoToBlob(info *sqlblobs.WorkflowExecutionInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ActivityInfoToBlob(info *sqlblobs.ActivityInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ChildExecutionInfoToBlob(info *sqlblobs.ChildExecutionInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) SignalInfoToBlob(info *sqlblobs.SignalInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) RequestCancelInfoToBlob(info *sqlblobs.RequestCancelInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TimerInfoToBlob(info *sqlblobs.TimerInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TaskInfoToBlob(info *sqlblobs.TaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TaskListInfoToBlob(info *sqlblobs.TaskListInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TransferTaskInfoToBlob(info *sqlblobs.TransferTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) TimerTaskInfoToBlob(info *sqlblobs.TimerTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func (e *thriftEncoder) ReplicationTaskInfoToBlob(info *sqlblobs.ReplicationTaskInfo) ([]byte, error) {
	return thriftRWEncode(info)
}

func thriftRWEncode(t thriftRWType) ([]byte, error) {
	value, err := t.ToWire()
	if err != nil {
		return nil, encodeErr(err)
	}
	var b bytes.Buffer
	if err := protocol.Binary.Encode(value, &b); err != nil {
		return nil, encodeErr(err)
	}
	return b.Bytes(), nil
}

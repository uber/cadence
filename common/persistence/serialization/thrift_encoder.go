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
	"github.com/uber/cadence/.gen/go/sqlblobs"
	p "github.com/uber/cadence/common/persistence"
)

type thriftEncoder struct{}

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

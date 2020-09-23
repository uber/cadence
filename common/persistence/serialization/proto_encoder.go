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
)

type protoEncoder struct{}

// NewProtoEncoder returns a new proto encoder
func NewProtoEncoder() Encoder {
	return &protoEncoder{}
}

func (e *protoEncoder) ShardInfoToBlob(*sqlblobs.ShardInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) DomainInfoToBlob(*sqlblobs.DomainInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) ActivityInfoToBlob(*sqlblobs.ActivityInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) SignalInfoToBlob(*sqlblobs.SignalInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) TimerInfoToBlob(*sqlblobs.TimerInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) TaskInfoToBlob(*sqlblobs.TaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) TaskListInfoToBlob(*sqlblobs.TaskListInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) ([]byte, error) {
	panic("not implemented")
}

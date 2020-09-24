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
	"github.com/uber/cadence/common"
)

type protoEncoder struct{}

func newProtoEncoder() encoder {
	return &protoEncoder{}
}

func (e *protoEncoder) shardInfoToBlob(*sqlblobs.ShardInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) domainInfoToBlob(*sqlblobs.DomainInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) historyTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) workflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) activityInfoToBlob(*sqlblobs.ActivityInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) childExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) signalInfoToBlob(*sqlblobs.SignalInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) requestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) timerInfoToBlob(*sqlblobs.TimerInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) taskInfoToBlob(*sqlblobs.TaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) taskListInfoToBlob(*sqlblobs.TaskListInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) transferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) timerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) replicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) encodingType() common.EncodingType {
	return common.EncodingTypeProto
}

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
	"github.com/uber/cadence/common"
)

type protoEncoder struct{}

func newProtoEncoder() encoder {
	return &protoEncoder{}
}

func (e *protoEncoder) shardInfoToBlob(*ShardInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) domainInfoToBlob(*DomainInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) historyTreeInfoToBlob(*HistoryTreeInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) workflowExecutionInfoToBlob(*WorkflowExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) activityInfoToBlob(*ActivityInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) childExecutionInfoToBlob(*ChildExecutionInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) signalInfoToBlob(*SignalInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) requestCancelInfoToBlob(*RequestCancelInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) timerInfoToBlob(*TimerInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) taskInfoToBlob(*TaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) taskListInfoToBlob(*TaskListInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) transferTaskInfoToBlob(*TransferTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) crossClusterTaskInfoToBlob(*CrossClusterTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) timerTaskInfoToBlob(*TimerTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) replicationTaskInfoToBlob(*ReplicationTaskInfo) ([]byte, error) {
	panic("not implemented")
}

func (e *protoEncoder) encodingType() common.EncodingType {
	return common.EncodingTypeProto
}

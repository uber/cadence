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

type (
	protoDecoder struct{}
)

func newProtoDecoder() decoder {
	return &protoDecoder{}
}

func (d *protoDecoder) shardInfoFromBlob(data []byte) (*sqlblobs.ShardInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) domainInfoFromBlob(data []byte) (*sqlblobs.DomainInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) historyTreeInfoFromBlob(data []byte) (*sqlblobs.HistoryTreeInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) workflowExecutionInfoFromBlob(data []byte) (*sqlblobs.WorkflowExecutionInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) activityInfoFromBlob(data []byte) (*sqlblobs.ActivityInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) childExecutionInfoFromBlob(data []byte) (*sqlblobs.ChildExecutionInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) signalInfoFromBlob(data []byte) (*sqlblobs.SignalInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) requestCancelInfoFromBlob(data []byte) (*sqlblobs.RequestCancelInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) timerInfoFromBlob(data []byte) (*sqlblobs.TimerInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) taskInfoFromBlob(data []byte) (*sqlblobs.TaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) taskListInfoFromBlob(data []byte) (*sqlblobs.TaskListInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) transferTaskInfoFromBlob(data []byte) (*sqlblobs.TransferTaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) timerTaskInfoFromBlob(data []byte) (*sqlblobs.TimerTaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) replicationTaskInfoFromBlob(data []byte) (*sqlblobs.ReplicationTaskInfo, error) {
	panic("not implemented")
}

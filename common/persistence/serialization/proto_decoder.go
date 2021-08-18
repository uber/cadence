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

type (
	protoDecoder struct{}
)

func newProtoDecoder() decoder {
	return &protoDecoder{}
}

func (d *protoDecoder) shardInfoFromBlob(data []byte) (*ShardInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) domainInfoFromBlob(data []byte) (*DomainInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) historyTreeInfoFromBlob(data []byte) (*HistoryTreeInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) workflowExecutionInfoFromBlob(data []byte) (*WorkflowExecutionInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) activityInfoFromBlob(data []byte) (*ActivityInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) childExecutionInfoFromBlob(data []byte) (*ChildExecutionInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) signalInfoFromBlob(data []byte) (*SignalInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) requestCancelInfoFromBlob(data []byte) (*RequestCancelInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) timerInfoFromBlob(data []byte) (*TimerInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) taskInfoFromBlob(data []byte) (*TaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) taskListInfoFromBlob(data []byte) (*TaskListInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) transferTaskInfoFromBlob(data []byte) (*TransferTaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) crossClusterTaskInfoFromBlob(data []byte) (*CrossClusterTaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) timerTaskInfoFromBlob(data []byte) (*TimerTaskInfo, error) {
	panic("not implemented")
}

func (d *protoDecoder) replicationTaskInfoFromBlob(data []byte) (*ReplicationTaskInfo, error) {
	panic("not implemented")
}

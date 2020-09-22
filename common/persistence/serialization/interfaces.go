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

type (
	// Encoder is used to serialize to persistence data blobs
	Encoder interface {
		ShardInfoToBlob(*sqlblobs.ShardInfo) (p.DataBlob, error)
		DomainInfoToBlob(*sqlblobs.DomainInfo) (p.DataBlob, error)
		HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) (p.DataBlob, error)
		WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) (p.DataBlob, error)
		ActivityInfoToBlob(*sqlblobs.ActivityInfo) (p.DataBlob, error)
		ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) (p.DataBlob, error)
		SignalInfoToBlob(*sqlblobs.SignalInfo) (p.DataBlob, error)
		RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) (p.DataBlob, error)
		TimerInfoToBlob(*sqlblobs.TimerInfo) (p.DataBlob, error)
		TaskInfoToBlob(*sqlblobs.TaskInfo) (p.DataBlob, error)
		TaskListInfoToBlob(*sqlblobs.TaskListInfo) (p.DataBlob, error)
		TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) (p.DataBlob, error)
		TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) (p.DataBlob, error)
		ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) (p.DataBlob, error)
	}

	// Decoder is used to deserialize from persistence data blobs
	Decoder interface {
		ShardInfoFromBlob([]byte, string) (*sqlblobs.ShardInfo, error)
		DomainInfoFromBlob([]byte, string) (*sqlblobs.DomainInfo, error)
		HistoryTreeInfoFromBlob([]byte, string) (*sqlblobs.HistoryTreeInfo, error)
		WorkflowExecutionInfoFromBlob([]byte, string) (*sqlblobs.WorkflowExecutionInfo, error)
		ActivityInfoFromBlob([]byte, string) (*sqlblobs.ActivityInfo, error)
		ChildExecutionInfoFromBlob([]byte, string) (*sqlblobs.ChildExecutionInfo, error)
		SignalInfoFromBlob([]byte, string) (*sqlblobs.SignalInfo, error)
		RequestCancelInfoFromBlob([]byte, string) (*sqlblobs.RequestCancelInfo, error)
		TimerInfoFromBlob([]byte, string) (*sqlblobs.TimerInfo, error)
		TaskInfoFromBlob([]byte, string) (*sqlblobs.TaskInfo, error)
		TaskListInfoFromBlob(b []byte, proto string) (*sqlblobs.TaskListInfo, error)
		TransferTaskInfoFromBlob([]byte, string) (*sqlblobs.TransferTaskInfo, error)
		TimerTaskInfoFromBlob([]byte, string) (*sqlblobs.TimerTaskInfo, error)
		ReplicationTaskInfoFromBlob([]byte, string) (*sqlblobs.ReplicationTaskInfo, error)
	}
)

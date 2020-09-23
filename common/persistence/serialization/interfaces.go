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
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/thriftrw/wire"
)

type (
	// Parser is used to do serialization and deserialization. A parser is backed by a
	// a single encoder which encodes into one format and a collection of decoders.
	// Parser selects the appropriate decoder for the provided blob.
	Parser interface {
		ShardInfoToBlob(*sqlblobs.ShardInfo) (persistence.DataBlob, error)
		DomainInfoToBlob(*sqlblobs.DomainInfo) (persistence.DataBlob, error)
		HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) (persistence.DataBlob, error)
		WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) (persistence.DataBlob, error)
		ActivityInfoToBlob(*sqlblobs.ActivityInfo) (persistence.DataBlob, error)
		ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) (persistence.DataBlob, error)
		SignalInfoToBlob(*sqlblobs.SignalInfo) (persistence.DataBlob, error)
		RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) (persistence.DataBlob, error)
		TimerInfoToBlob(*sqlblobs.TimerInfo) (persistence.DataBlob, error)
		TaskInfoToBlob(*sqlblobs.TaskInfo) (persistence.DataBlob, error)
		TaskListInfoToBlob(*sqlblobs.TaskListInfo) (persistence.DataBlob, error)
		TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) (persistence.DataBlob, error)
		TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) (persistence.DataBlob, error)
		ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) (persistence.DataBlob, error)

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
		TaskListInfoFromBlob([]byte, string) (*sqlblobs.TaskListInfo, error)
		TransferTaskInfoFromBlob([]byte, string) (*sqlblobs.TransferTaskInfo, error)
		TimerTaskInfoFromBlob([]byte, string) (*sqlblobs.TimerTaskInfo, error)
		ReplicationTaskInfoFromBlob([]byte, string) (*sqlblobs.ReplicationTaskInfo, error)
	}

	// Encoder is used to serialize structs. Each encoder implementation uses one serialization format.
	Encoder interface {
		ShardInfoToBlob(*sqlblobs.ShardInfo) ([]byte, error)
		DomainInfoToBlob(*sqlblobs.DomainInfo) ([]byte, error)
		HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) ([]byte, error)
		WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) ([]byte, error)
		ActivityInfoToBlob(*sqlblobs.ActivityInfo) ([]byte, error)
		ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) ([]byte, error)
		SignalInfoToBlob(*sqlblobs.SignalInfo) ([]byte, error)
		RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) ([]byte, error)
		TimerInfoToBlob(*sqlblobs.TimerInfo) ([]byte, error)
		TaskInfoToBlob(*sqlblobs.TaskInfo) ([]byte, error)
		TaskListInfoToBlob(*sqlblobs.TaskListInfo) ([]byte, error)
		TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) ([]byte, error)
		TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) ([]byte, error)
		ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) ([]byte, error)
	}

	// Decoder is used to deserialize structs. Each decoder implementation uses one serialization format.
	Decoder interface {
		ShardInfoFromBlob([]byte) (*sqlblobs.ShardInfo, error)
		DomainInfoFromBlob([]byte) (*sqlblobs.DomainInfo, error)
		HistoryTreeInfoFromBlob([]byte) (*sqlblobs.HistoryTreeInfo, error)
		WorkflowExecutionInfoFromBlob([]byte) (*sqlblobs.WorkflowExecutionInfo, error)
		ActivityInfoFromBlob([]byte) (*sqlblobs.ActivityInfo, error)
		ChildExecutionInfoFromBlob([]byte) (*sqlblobs.ChildExecutionInfo, error)
		SignalInfoFromBlob([]byte) (*sqlblobs.SignalInfo, error)
		RequestCancelInfoFromBlob([]byte) (*sqlblobs.RequestCancelInfo, error)
		TimerInfoFromBlob([]byte) (*sqlblobs.TimerInfo, error)
		TaskInfoFromBlob([]byte) (*sqlblobs.TaskInfo, error)
		TaskListInfoFromBlob([]byte) (*sqlblobs.TaskListInfo, error)
		TransferTaskInfoFromBlob([]byte) (*sqlblobs.TransferTaskInfo, error)
		TimerTaskInfoFromBlob([]byte) (*sqlblobs.TimerTaskInfo, error)
		ReplicationTaskInfoFromBlob([]byte) (*sqlblobs.ReplicationTaskInfo, error)
	}

	thriftRWType interface {
		ToWire() (wire.Value, error)
		FromWire(w wire.Value) error
	}
)

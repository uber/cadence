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
	"go.uber.org/thriftrw/wire"

	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
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

	// encoder is used to serialize structs. Each encoder implementation uses one serialization format.
	encoder interface {
		shardInfoToBlob(*sqlblobs.ShardInfo) ([]byte, error)
		domainInfoToBlob(*sqlblobs.DomainInfo) ([]byte, error)
		historyTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) ([]byte, error)
		workflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) ([]byte, error)
		activityInfoToBlob(*sqlblobs.ActivityInfo) ([]byte, error)
		childExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) ([]byte, error)
		signalInfoToBlob(*sqlblobs.SignalInfo) ([]byte, error)
		requestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) ([]byte, error)
		timerInfoToBlob(*sqlblobs.TimerInfo) ([]byte, error)
		taskInfoToBlob(*sqlblobs.TaskInfo) ([]byte, error)
		taskListInfoToBlob(*sqlblobs.TaskListInfo) ([]byte, error)
		transferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) ([]byte, error)
		timerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) ([]byte, error)
		replicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) ([]byte, error)
		encodingType() common.EncodingType
	}

	// decoder is used to deserialize structs. Each decoder implementation uses one serialization format.
	decoder interface {
		shardInfoFromBlob([]byte) (*sqlblobs.ShardInfo, error)
		domainInfoFromBlob([]byte) (*sqlblobs.DomainInfo, error)
		historyTreeInfoFromBlob([]byte) (*sqlblobs.HistoryTreeInfo, error)
		workflowExecutionInfoFromBlob([]byte) (*sqlblobs.WorkflowExecutionInfo, error)
		activityInfoFromBlob([]byte) (*sqlblobs.ActivityInfo, error)
		childExecutionInfoFromBlob([]byte) (*sqlblobs.ChildExecutionInfo, error)
		signalInfoFromBlob([]byte) (*sqlblobs.SignalInfo, error)
		requestCancelInfoFromBlob([]byte) (*sqlblobs.RequestCancelInfo, error)
		timerInfoFromBlob([]byte) (*sqlblobs.TimerInfo, error)
		taskInfoFromBlob([]byte) (*sqlblobs.TaskInfo, error)
		taskListInfoFromBlob([]byte) (*sqlblobs.TaskListInfo, error)
		transferTaskInfoFromBlob([]byte) (*sqlblobs.TransferTaskInfo, error)
		timerTaskInfoFromBlob([]byte) (*sqlblobs.TimerTaskInfo, error)
		replicationTaskInfoFromBlob([]byte) (*sqlblobs.ReplicationTaskInfo, error)
	}

	thriftRWType interface {
		ToWire() (wire.Value, error)
		FromWire(w wire.Value) error
	}
)

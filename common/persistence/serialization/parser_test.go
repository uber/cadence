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
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

func TestParserRoundTrip(t *testing.T) {
	thriftParser, err := NewParser(common.EncodingTypeThriftRW, common.EncodingTypeThriftRW)
	assert.NoError(t, err)
	f := fuzz.New().Funcs(func(e *time.Time, c fuzz.Continue) {
		*e = time.Unix(c.Int63n(1000000), 0)
	}, func(e *time.Duration, c fuzz.Continue) {
		*e = time.Duration(common.DurationToDays(time.Duration(c.Int63n(1000000))))
	}, func(e *UUID, c fuzz.Continue) {
		*e = MustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	}).NilChance(0).NumElements(0, 10).
		// IsCron is dynamic fields that is connected to CronSchedule, so we need to skip it and test separately.
		SkipFieldsWithPattern(regexp.MustCompile("IsCron|CronSchedule"))

	for _, testCase := range []any{
		&ShardInfo{},
		&DomainInfo{},
		&HistoryTreeInfo{},
		&WorkflowExecutionInfo{},
		&ActivityInfo{},
		&ChildExecutionInfo{},
		&SignalInfo{},
		&RequestCancelInfo{},
		&TimerInfo{},
		&TaskInfo{},
		&TaskListInfo{},
		&TransferTaskInfo{},
		&TimerTaskInfo{},
		&ReplicationTaskInfo{},
	} {
		t.Run(reflect.TypeOf(testCase).String(), func(t *testing.T) {
			f.Fuzz(testCase)
			blob := parse(t, thriftParser, testCase)
			result := unparse(t, thriftParser, blob, testCase)
			assert.Equal(t, testCase, result)
		})
	}
}

func parse(t *testing.T, parser Parser, data any) persistence.DataBlob {
	var (
		blob persistence.DataBlob
		err  error
	)
	switch v := data.(type) {
	case *ShardInfo:
		blob, err = parser.ShardInfoToBlob(v)
	case *DomainInfo:
		blob, err = parser.DomainInfoToBlob(v)
	case *HistoryTreeInfo:
		blob, err = parser.HistoryTreeInfoToBlob(v)
	case *WorkflowExecutionInfo:
		blob, err = parser.WorkflowExecutionInfoToBlob(v)
	case *ActivityInfo:
		blob, err = parser.ActivityInfoToBlob(v)
	case *ChildExecutionInfo:
		blob, err = parser.ChildExecutionInfoToBlob(v)
	case *SignalInfo:
		blob, err = parser.SignalInfoToBlob(v)
	case *RequestCancelInfo:
		blob, err = parser.RequestCancelInfoToBlob(v)
	case *TimerInfo:
		blob, err = parser.TimerInfoToBlob(v)
	case *TaskInfo:
		blob, err = parser.TaskInfoToBlob(v)
	case *TaskListInfo:
		blob, err = parser.TaskListInfoToBlob(v)
	case *TransferTaskInfo:
		blob, err = parser.TransferTaskInfoToBlob(v)
	case *TimerTaskInfo:
		blob, err = parser.TimerTaskInfoToBlob(v)
	case *ReplicationTaskInfo:
		blob, err = parser.ReplicationTaskInfoToBlob(v)
	default:
		err = fmt.Errorf("unknown type %T", v)
	}
	assert.NoError(t, err)
	return blob
}

func unparse(t *testing.T, parser Parser, blob persistence.DataBlob, result any) any {
	var (
		data any
		err  error
	)
	switch v := result.(type) {
	case *ShardInfo:
		data, err = parser.ShardInfoFromBlob(blob.Data, string(blob.Encoding))
	case *DomainInfo:
		data, err = parser.DomainInfoFromBlob(blob.Data, string(blob.Encoding))
	case *HistoryTreeInfo:
		data, err = parser.HistoryTreeInfoFromBlob(blob.Data, string(blob.Encoding))
	case *WorkflowExecutionInfo:
		data, err = parser.WorkflowExecutionInfoFromBlob(blob.Data, string(blob.Encoding))
	case *ActivityInfo:
		data, err = parser.ActivityInfoFromBlob(blob.Data, string(blob.Encoding))
	case *ChildExecutionInfo:
		data, err = parser.ChildExecutionInfoFromBlob(blob.Data, string(blob.Encoding))
	case *SignalInfo:
		data, err = parser.SignalInfoFromBlob(blob.Data, string(blob.Encoding))
	case *RequestCancelInfo:
		data, err = parser.RequestCancelInfoFromBlob(blob.Data, string(blob.Encoding))
	case *TimerInfo:
		data, err = parser.TimerInfoFromBlob(blob.Data, string(blob.Encoding))
	case *TaskInfo:
		data, err = parser.TaskInfoFromBlob(blob.Data, string(blob.Encoding))
	case *TaskListInfo:
		data, err = parser.TaskListInfoFromBlob(blob.Data, string(blob.Encoding))
	case *TransferTaskInfo:
		data, err = parser.TransferTaskInfoFromBlob(blob.Data, string(blob.Encoding))
	case *TimerTaskInfo:
		data, err = parser.TimerTaskInfoFromBlob(blob.Data, string(blob.Encoding))
	case *ReplicationTaskInfo:
		data, err = parser.ReplicationTaskInfoFromBlob(blob.Data, string(blob.Encoding))
	default:
		err = fmt.Errorf("unknown type %T", v)
	}
	assert.NoError(t, err)
	return data
}

func TestParser_WorkflowExecution_with_cron(t *testing.T) {
	info := &WorkflowExecutionInfo{
		CronSchedule: "@every 1m",
		IsCron:       true,
	}
	parser, err := NewParser(common.EncodingTypeThriftRW, common.EncodingTypeThriftRW)
	require.NoError(t, err)
	blob, err := parser.WorkflowExecutionInfoToBlob(info)
	require.NoError(t, err)
	result, err := parser.WorkflowExecutionInfoFromBlob(blob.Data, string(blob.Encoding))
	require.NoError(t, err)
	assert.Equal(t, info, result)
}

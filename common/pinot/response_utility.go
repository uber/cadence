// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package pinot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	Memo         = "Memo"
	MemoEncoding = "Memo_Encoding"
)

func buildMap(hit []interface{}, columnNames []string) map[string]interface{} {
	systemKeyMap := make(map[string]interface{})

	for i := 0; i < len(columnNames); i++ {
		key := columnNames[i]
		systemKeyMap[key] = hit[i]
	}

	return systemKeyMap
}

// VisibilityRecord is a struct of doc for deserialization
// this is different from InternalVisibilityWorkflowExecutionInfo
// use this to deserialize the systemKeyMap from Pinot response
type VisibilityRecord struct {
	WorkflowID    string
	RunID         string
	WorkflowType  string
	DomainID      string
	StartTime     int64
	ExecutionTime int64
	CloseTime     int64
	CloseStatus   int
	HistoryLength int64
	TaskList      string
	IsCron        bool
	NumClusters   int16
	UpdateTime    int64
	ShardID       int16
}

func ConvertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string, logger log.Logger) *p.InternalVisibilityWorkflowExecutionInfo {
	if len(hit) != len(columnNames) {
		return nil
	}

	systemKeyMap := buildMap(hit, columnNames)

	jsonSystemKeyMap, err := json.Marshal(systemKeyMap)
	if err != nil {
		logger.Error("unable to marshal systemKeyMap",
			tag.Error(err))
		return nil
	}

	attributeMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(fmt.Sprint(systemKeyMap["Attr"])), &attributeMap)
	if err != nil {
		logger.Error("Unable to Unmarshal searchAttribute map", tag.Error(err))
		return nil
	}

	var memo *p.DataBlob
	if attributeMap[Memo] != nil {
		memo, err = convertMemo(attributeMap[Memo], attributeMap[MemoEncoding])
	}
	delete(attributeMap, Memo)         // cleanup after we get memo from search attribute
	delete(attributeMap, MemoEncoding) // cleanup after we get memo from search attribute

	var source *VisibilityRecord
	err = json.Unmarshal(jsonSystemKeyMap, &source)
	if err != nil {
		logger.Error("Unable to Unmarshal systemKeyMap",
			tag.Error(err),
		)
		return nil
	}

	record := &p.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         source.DomainID,
		WorkflowType:     source.WorkflowType,
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.UnixMilli(source.StartTime), // be careful: source.StartTime is in milliseconds
		ExecutionTime:    time.UnixMilli(source.ExecutionTime),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		NumClusters:      source.NumClusters,
		ShardID:          source.ShardID,
		SearchAttributes: attributeMap,
		Memo:             memo,
	}
	if source.UpdateTime > 0 {
		record.UpdateTime = time.UnixMilli(source.UpdateTime)
	}
	if source.CloseTime > 0 {
		record.CloseTime = time.UnixMilli(source.CloseTime)
		record.Status = toWorkflowExecutionCloseStatus(source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func convertMemo(data interface{}, encoding interface{}) (*p.DataBlob, error) {
	memoData, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("couldn't convert data to byte array. %T", data)
	}
	if memoData == nil || len(memoData) == 0 || string(memoData) == "null" {
		return nil, nil
	}

	memoEncoding, ok := encoding.(common.EncodingType)
	if !ok {
		return nil, fmt.Errorf("couldn't convert encoding to string. %T", encoding)
	}

	if memoEncoding != common.EncodingTypeJSON {
		return nil, fmt.Errorf("encoding type doesn't support. Looking for EncodingTypeJSON, but found %v", encoding)
	}

	var res p.DataBlob
	err := json.Unmarshal(memoData, &res)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func toWorkflowExecutionCloseStatus(status int) *types.WorkflowExecutionCloseStatus {
	if status < 0 {
		return nil
	}
	closeStatus := types.WorkflowExecutionCloseStatus(status)
	return &closeStatus
}

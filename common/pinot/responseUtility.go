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
	"time"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func buildMap(hit []interface{}, columnNames []string, isSystemKey func(key string) bool) (map[string]interface{}, map[string]interface{}) {
	systemKeyMap := make(map[string]interface{})
	customKeyMap := make(map[string]interface{})

	for i := 0; i < len(columnNames); i++ {
		key := columnNames[i]
		// checks if it is system key, if yes, put it into the system map; otherwise put it into custom map
		if isSystemKey(key) {
			systemKeyMap[key] = hit[i]
		} else {
			customKeyMap[key] = hit[i]
		}
	}

	return systemKeyMap, customKeyMap
}

// VisibilityRecord is a struct of doc for deserialization
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
	Encoding      string
	TaskList      string
	IsCron        bool
	NumClusters   int16
	UpdateTime    int64
	Attr          string
}

func ConvertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string, logger log.Logger, isSystemKey func(key string) bool) *p.InternalVisibilityWorkflowExecutionInfo {
	if len(hit) != len(columnNames) {
		return nil
	}

	systemKeyMap, customKeyMap := buildMap(hit, columnNames, isSystemKey)
	jsonSystemKeyMap, err := json.Marshal(systemKeyMap)
	if err != nil { // log and skip error
		logger.Error("unable to marshal systemKeyMap",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	var source *VisibilityRecord
	err = json.Unmarshal(jsonSystemKeyMap, &source)
	if err != nil { // log and skip error
		logger.Error("unable to Unmarshal systemKeyMap",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	record := &p.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         source.DomainID,
		WorkflowType:     source.WorkflowType,
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.UnixMilli(source.StartTime), // be careful: source.StartTime is in milisecond
		ExecutionTime:    time.UnixMilli(source.ExecutionTime),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		NumClusters:      source.NumClusters,
		SearchAttributes: customKeyMap,
	}
	if source.UpdateTime != 0 {
		record.UpdateTime = time.UnixMilli(source.UpdateTime)
	}
	if source.CloseTime != 0 {
		record.CloseTime = time.UnixMilli(source.CloseTime)
		record.Status = toWorkflowExecutionCloseStatus(source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func toWorkflowExecutionCloseStatus(status int) *types.WorkflowExecutionCloseStatus {
	if status < 0 {
		return nil
	}
	closeStatus := types.WorkflowExecutionCloseStatus(status)
	return &closeStatus
}

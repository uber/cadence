// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package testdata

import (
	"time"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

const (
	DomainID      = "test-domain-id"
	WorkflowType  = "test-workflow-type"
	WorkflowID    = "test-workflow-id"
	RunID         = "test-run-id"
	TypeName      = "test-type-name"
	HistoryLenght = int64(1)
	TaskList      = "test-task-list"
	NumClusters   = int16(1)
	ShardID       = int16(1)
)

var CurrentTime = time.Now()

func NewVisibilityRow() persistence.InternalVisibilityWorkflowExecutionInfo {
	return persistence.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         DomainID,
		WorkflowType:     WorkflowType,
		WorkflowID:       WorkflowID,
		RunID:            RunID,
		TypeName:         TypeName,
		StartTime:        CurrentTime,
		ExecutionTime:    CurrentTime,
		CloseTime:        CurrentTime,
		Status:           types.WorkflowExecutionCloseStatusCompleted.Ptr(),
		HistoryLength:    HistoryLenght,
		Memo:             &persistence.DataBlob{},
		TaskList:         TaskList,
		IsCron:           false,
		NumClusters:      NumClusters,
		UpdateTime:       CurrentTime,
		SearchAttributes: map[string]interface{}{},
		ShardID:          ShardID,
	}
}

func NewVisibilityRowForInsert() *nosqlplugin.VisibilityRowForInsert {
	return &nosqlplugin.VisibilityRowForInsert{
		VisibilityRow: NewVisibilityRow(),
		DomainID:      DomainID,
	}
}

func NewVisibilityRowForUpdate(updateCloseToOpen, updateOpenToClose bool) *nosqlplugin.VisibilityRowForUpdate {
	visibilityRow := NewVisibilityRow()
	visibilityRow.CloseTime = visibilityRow.StartTime.Add(-1 * time.Minute)
	return &nosqlplugin.VisibilityRowForUpdate{
		VisibilityRow:     visibilityRow,
		DomainID:          DomainID,
		UpdateCloseToOpen: updateCloseToOpen,
		UpdateOpenToClose: updateOpenToClose,
	}
}

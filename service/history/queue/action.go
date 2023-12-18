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

package queue

import "github.com/uber/cadence/common/types"

type (
	// ActionType specifies the type of the Action
	ActionType int

	// Action specifies the Action should be performed
	Action struct {
		ActionType               ActionType
		ResetActionAttributes    *ResetActionAttributes
		GetStateActionAttributes *GetStateActionAttributes
		GetTasksAttributes       *GetTasksAttributes
		UpdateTaskAttributes     *UpdateTasksAttributes
		// add attributes for other action types here
	}

	// ActionResult is the result for performing an Action
	ActionResult struct {
		ActionType           ActionType
		ResetActionResult    *ResetActionResult
		GetStateActionResult *GetStateActionResult
		GetTasksResult       *GetTasksResult
		UpdateTaskResult     *UpdateTasksResult
	}

	// ResetActionAttributes contains the parameter for performing Reset Action
	ResetActionAttributes struct{}
	// ResetActionResult is the result for performing Reset Action
	ResetActionResult struct{}

	// GetStateActionAttributes contains the parameter for performing GetState Action
	GetStateActionAttributes struct{}
	// GetStateActionResult is the result for performing GetState Action
	GetStateActionResult struct {
		States []ProcessingQueueState
	}

	// GetTasksAttributes contains the parameter to get tasks
	GetTasksAttributes struct{}
	// GetTasksResult is the result for performing GetTasks Action
	GetTasksResult struct {
		TaskRequests []*types.CrossClusterTaskRequest
	}
	// UpdateTasksAttributes contains the parameter to update task
	UpdateTasksAttributes struct {
		TaskResponses []*types.CrossClusterTaskResponse
	}
	// UpdateTasksResult is the result for performing UpdateTask Action
	UpdateTasksResult struct {
	}
)

const (
	// ActionTypeReset is the ActionType for reseting processing queue states
	ActionTypeReset ActionType = iota + 1
	// ActionTypeGetState is the ActionType for reading processing queue states
	ActionTypeGetState
	// ActionTypeGetTasks is the ActionType for get cross cluster tasks
	ActionTypeGetTasks
	// ActionTypeUpdateTask is the ActionType to update outstanding task
	ActionTypeUpdateTask
	// add more ActionType here
)

// NewResetAction creates a new action for reseting processing queue states
func NewResetAction() *Action {
	return &Action{
		ActionType:            ActionTypeReset,
		ResetActionAttributes: &ResetActionAttributes{},
	}
}

// NewGetStateAction reads all processing queue states in the processor
func NewGetStateAction() *Action {
	return &Action{
		ActionType:               ActionTypeGetState,
		GetStateActionAttributes: &GetStateActionAttributes{},
	}
}

// NewGetTasksAction creates a queue action for fetching cross cluster tasks
func NewGetTasksAction() *Action {
	return &Action{
		ActionType:         ActionTypeGetTasks,
		GetTasksAttributes: &GetTasksAttributes{},
	}
}

// NewUpdateTasksAction creates a queue action for responding cross cluster task
// processing results
func NewUpdateTasksAction(
	taskResponses []*types.CrossClusterTaskResponse,
) *Action {
	return &Action{
		ActionType: ActionTypeUpdateTask,
		UpdateTaskAttributes: &UpdateTasksAttributes{
			TaskResponses: taskResponses,
		},
	}
}

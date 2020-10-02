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

package thrift

import (
	thrift "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

// FromTaskListKind converts internal TaskListKind type to thrift
func FromTaskListKind(kind types.TaskListKind) *thrift.TaskListKind {
	switch kind {
	case types.TaskListKindNormal:
		return taskListKindPtr(thrift.TaskListKindNormal)
	case types.TaskListKindSticky:
		return taskListKindPtr(thrift.TaskListKindSticky)
	}

	panic("unexpected TaskListKind value")
}

// ToTaskListKind converts thrift TaskListKind type to internal
func ToTaskListKind(kind *thrift.TaskListKind) types.TaskListKind {
	if kind == nil {
		return types.TaskListKindNormal
	}

	switch *kind {
	case thrift.TaskListKindNormal:
		return types.TaskListKindNormal
	case thrift.TaskListKindSticky:
		return types.TaskListKindSticky
	}

	panic("unexpected TaskListKind value")
}

// FromTaskList converts internal TaskList type to thrift
func FromTaskList(tl types.TaskList) *thrift.TaskList {
	return &thrift.TaskList{
		Name: stringPtr(tl.Name),
		Kind: FromTaskListKind(tl.Kind),
	}
}

// ToTaskList converts thrift TaskList type to internal
func ToTaskList(tl *thrift.TaskList) types.TaskList {
	if tl == nil {
		return types.TaskList{}
	}
	return types.TaskList{
		Name: stringVal(tl.Name),
		Kind: ToTaskListKind(tl.Kind),
	}
}

func taskListKindPtr(kind thrift.TaskListKind) *thrift.TaskListKind {
	return &kind
}

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
	"testing"

	"github.com/stretchr/testify/assert"
	thrift "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

func Test_FromTaskListKind(t *testing.T) {
	assert.Equal(t, thrift.TaskListKindNormal.Ptr(), FromTaskListKind(types.TaskListKindNormal))
	assert.Equal(t, thrift.TaskListKindSticky.Ptr(), FromTaskListKind(types.TaskListKindSticky))
	assert.Equal(t, thrift.TaskListKind(99).Ptr(), FromTaskListKind(types.TaskListKind(99)))
}

func Test_ToTaskListKind(t *testing.T) {
	assert.Equal(t, types.TaskListKindNormal, ToTaskListKind(nil))
	assert.Equal(t, types.TaskListKindNormal, ToTaskListKind(thrift.TaskListKindNormal.Ptr()))
	assert.Equal(t, types.TaskListKindSticky, ToTaskListKind(thrift.TaskListKindSticky.Ptr()))
	assert.Equal(t, types.TaskListKind(99), ToTaskListKind(thrift.TaskListKind(99).Ptr()))
}

func Test_FromTaskList(t *testing.T) {
	assert.Equal(t,
		&thrift.TaskList{
			Name: stringPtr(""),
			Kind: thrift.TaskListKindNormal.Ptr(),
		},
		FromTaskList(types.TaskList{}),
	)
	assert.Equal(t,
		&thrift.TaskList{
			Name: stringPtr("Name"),
			Kind: thrift.TaskListKindSticky.Ptr(),
		},
		FromTaskList(types.TaskList{
			Name: "Name",
			Kind: types.TaskListKindSticky,
		}),
	)
}

func Test_ToTaskList(t *testing.T) {
	assert.Equal(t,
		types.TaskList{},
		ToTaskList(nil),
	)
	assert.Equal(t,
		types.TaskList{},
		ToTaskList(&thrift.TaskList{}),
	)
	assert.Equal(t,
		types.TaskList{
			Name: "Name",
			Kind: types.TaskListKindSticky,
		},
		ToTaskList(&thrift.TaskList{
			Name: stringPtr("Name"),
			Kind: thrift.TaskListKindSticky.Ptr(),
		}),
	)
}

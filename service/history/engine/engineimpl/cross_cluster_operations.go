// Copyright (c) 2017-2021 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2021 Temporal Technologies Inc.
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

package engineimpl

import (
	"context"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/queue"
)

func (e *historyEngineImpl) GetCrossClusterTasks(
	ctx context.Context,
	targetCluster string,
) ([]*types.CrossClusterTaskRequest, error) {
	actionResult, err := e.crossClusterProcessor.HandleAction(ctx, targetCluster, queue.NewGetTasksAction())
	if err != nil {
		return nil, err
	}

	return actionResult.GetTasksResult.TaskRequests, nil
}

func (e *historyEngineImpl) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	targetCluster string,
	responses []*types.CrossClusterTaskResponse,
) error {
	_, err := e.crossClusterProcessor.HandleAction(ctx, targetCluster, queue.NewUpdateTasksAction(responses))
	return err
}

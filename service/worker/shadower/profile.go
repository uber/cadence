// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package shadower

import (
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/.gen/go/shadower"
)

type (
	workflowProfile struct {
		ctx       workflow.Context
		startTime time.Time
		scope     tally.Scope
		logger    *zap.Logger
	}
)

const (
	tagShadowDomain   = "shadow-domain"
	tagShadowTaskList = "shadow-tasklist"
)

const (
	shadowWorkflowLatency       = "shadow-workflow-latency"
	shadowWorkflowStarted       = "shadow-workflow-started"
	shadowWorkflowCompleted     = "shadow-workflow-completed"
	shadowWorkflowContinueAsNew = "shadow-workflow-continueasnew"
	shadowWorkflowFailed        = "shadow-workflow-failed"
)

func beginWorkflow(
	ctx workflow.Context,
	params *shadower.WorkflowParams,
) *workflowProfile {
	taggedScope := workflow.GetMetricsScope(ctx).Tagged(map[string]string{
		tagShadowDomain:   params.GetDomain(),
		tagShadowTaskList: params.GetTaskList(),
	})
	taggedLogger := workflow.GetLogger(ctx).With(
		zap.String(tagShadowDomain, params.GetDomain()),
		zap.String(tagShadowTaskList, params.GetTaskList()),
	)
	if params.LastRunResult == nil {
		taggedScope.Counter(shadowWorkflowStarted).Inc(1)
		taggedLogger.Info("Shadow workflow started")
	}
	return &workflowProfile{
		ctx:       ctx,
		startTime: workflow.Now(ctx),
		scope:     taggedScope,
		logger:    taggedLogger,
	}
}

func (p *workflowProfile) endWorkflow(
	err error,
) error {
	now := workflow.Now(p.ctx)
	p.scope.Timer(shadowWorkflowLatency).Record(now.Sub(p.startTime))
	switch err.(type) {
	case nil:
		p.scope.Counter(shadowWorkflowCompleted).Inc(1)
		p.logger.Info("Shadow workflow completed")
	case *workflow.ContinueAsNewError:
		p.scope.Counter(shadowWorkflowContinueAsNew).Inc(1)
		p.logger.Info("Shadow workflow continued as new")
	default:
		p.scope.Counter(shadowWorkflowFailed).Inc(1)
		p.logger.With(zap.Error(err)).Error("Shadow workflow failed")
	}
	return err
}

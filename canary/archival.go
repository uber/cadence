// Copyright (c) 2019 Uber Technologies, Inc.
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

package canary

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	numHistoryArchivals = 5
	resultSize          = 1024 // 1KB
)

const (
	// run visibility archival checks at a low probability
	// as each query can be expensive and take a long time
	visArchivalQueryProbability = 0.01
	visArchivalPageSize         = 100

	// TODO: remove the check on URI scheme once
	// https://github.com/uber/cadence/issues/3003 is done
	visArchivalScheme        = "file"
	visArchivalQueryTemplate = "WorkflowType = '%s' and CloseTime >= %v and CloseTime <= %v"
)

func init() {
	registerWorkflow(archivalWorkflow, wfTypeArchival)
	registerActivity(historyArchivalActivity, activityTypeHistoryArchival)
	registerWorkflow(historyArchivalExternalWorkflow, wfTypeHistoryArchivalExternal)
	registerActivity(largeResultActivity, activityTypeLargeResult)
	registerActivity(visibilityArchivalActivity, activityTypeVisibilityArchival)
}

func archivalWorkflow(ctx workflow.Context, scheduledTimeNanos int64, _ string) error {
	profile, err := beginWorkflow(ctx, wfTypeArchival, scheduledTimeNanos)
	if err != nil {
		return err
	}
	ch := workflow.NewBufferedChannel(ctx, numHistoryArchivals)
	for i := 0; i < numHistoryArchivals; i++ {
		workflow.Go(ctx, func(ctx2 workflow.Context) {
			aCtx := workflow.WithActivityOptions(ctx2, newActivityOptions())
			err := workflow.ExecuteActivity(aCtx, activityTypeHistoryArchival, workflow.Now(ctx2).UnixNano()).Get(aCtx, nil)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			ch.Send(ctx2, errStr)
		})
	}
	successfulArchivalsCount := 0
	for i := 0; i < numHistoryArchivals; i++ {
		var errStr string
		ch.Receive(ctx, &errStr)
		if errStr != "" {
			workflow.GetLogger(ctx).Error("at least one archival failed", zap.Int("success-count", successfulArchivalsCount), zap.String("err-string", errStr))
			return profile.end(errors.New(errStr))
		}
		successfulArchivalsCount++
	}

	var queryArchivedVisRecords bool
	if err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return rand.Float64() <= visArchivalQueryProbability
	}).Get(&queryArchivedVisRecords); err != nil {
		workflow.GetLogger(ctx).Error("unable to determine if visibility archival test should be run", zap.Error(err))
		return profile.end(err)
	}

	if queryArchivedVisRecords {
		aCtx := workflow.WithActivityOptions(ctx, newActivityOptionsWithRetry(nil))
		err := workflow.ExecuteActivity(aCtx, activityTypeVisibilityArchival, workflow.Now(ctx).UnixNano()).Get(aCtx, nil)
		if err != nil {
			workflow.GetLogger(aCtx).Error("failed to list archived workflows", zap.Error(err))
			return profile.end(err)
		}
	}

	return profile.end(nil)
}

func historyArchivalActivity(ctx context.Context, scheduledTimeNanos int64) error {
	scope := activity.GetMetricsScope(ctx)
	var err error
	scope, sw := recordActivityStart(scope, activityTypeHistoryArchival, scheduledTimeNanos)
	defer recordActivityEnd(scope, sw, err)

	client := getActivityArchivalContext(ctx).cadence
	workflowID := fmt.Sprintf("%v.%v", wfTypeHistoryArchivalExternal, uuid.New().String())
	ops := newWorkflowOptions(workflowID, childWorkflowTimeout)
	ops.TaskList = archivalTaskListName
	workflowRun, err := client.ExecuteWorkflow(context.Background(), ops, wfTypeHistoryArchivalExternal, scheduledTimeNanos)
	if err != nil {
		return err
	}
	err = workflowRun.Get(ctx, nil)
	if err != nil {
		return err
	}
	domain := archivalDomain
	runID := workflowRun.GetRunID()
	getHistoryReq := &shared.GetWorkflowExecutionHistoryRequest{
		Domain: &domain,
		Execution: &shared.WorkflowExecution{
			WorkflowId: &workflowID,
			RunId:      &runID,
		},
	}

	failureReason := ""
	attempts := 0
	expireTime := time.Now().Add(activityTaskTimeout)
	for {
		<-time.After(5 * time.Second)
		if time.Now().After(expireTime) {
			break
		}
		attempts++
		bCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		history, err := client.Service.GetWorkflowExecutionHistory(bCtx, getHistoryReq)
		cancel()
		if err != nil {
			failureReason = fmt.Sprintf("error accessing history, %v", err.Error())
		} else if !history.GetArchived() {
			failureReason = "history is not archived"
		} else if len(history.History.Events) == 0 {
			failureReason = "got empty history"
		} else {
			return nil
		}
	}
	activity.GetLogger(ctx).Error("failed to get archived history within time limit",
		zap.String("failure_reason", failureReason),
		zap.String("domain", archivalDomain),
		zap.String("workflow_id", workflowID),
		zap.String("run_id", runID),
		zap.Int("attempts", attempts))
	return fmt.Errorf("failed to get archived history within time limit, %v", failureReason)
}

func historyArchivalExternalWorkflow(ctx workflow.Context, scheduledTimeNanos int64) error {
	profile, err := beginWorkflow(ctx, wfTypeHistoryArchivalExternal, scheduledTimeNanos)
	if err != nil {
		return err
	}
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	var result []byte
	numActs := rand.Intn(10) + 1
	for i := 0; i != numActs; i++ {
		aCtx := workflow.WithActivityOptions(ctx, ao)
		if err = workflow.ExecuteActivity(aCtx, activityTypeLargeResult).Get(aCtx, &result); err != nil {
			break
		}
	}
	if err != nil {
		return profile.end(err)
	}
	return profile.end(nil)
}

func largeResultActivity() ([]byte, error) {
	return make([]byte, resultSize, resultSize), nil
}

func visibilityArchivalActivity(ctx context.Context, scheduledTimeNanos int64) error {
	scope := activity.GetMetricsScope(ctx)
	var err error
	scope, sw := recordActivityStart(scope, activityTypeVisibilityArchival, scheduledTimeNanos)
	defer recordActivityEnd(scope, sw, err)

	client := getActivityArchivalContext(ctx).cadence
	resp, err := client.Describe(ctx, archivalDomain)
	if err != nil {
		return err
	}

	if resp.Configuration != nil &&
		resp.Configuration.GetVisibilityArchivalStatus() == shared.ArchivalStatusDisabled {
		return errors.New("domain not configured for visibility archival")
	}

	visArchivalURI := ""
	if resp.Configuration != nil {
		visArchivalURI = resp.Configuration.GetVisibilityArchivalURI()
	}

	scheme := getURIScheme(visArchivalURI)
	if scheme != visArchivalScheme {
		return fmt.Errorf("unknown visibility archival scheme: %s, expecting %s", scheme, visArchivalScheme)
	}

	listReq := &shared.ListArchivedWorkflowExecutionsRequest{
		Domain:   stringPtr(archivalDomain),
		PageSize: int32Ptr(visArchivalPageSize),
		Query: stringPtr(fmt.Sprintf(visArchivalQueryTemplate,
			wfTypeArchival,
			time.Now().Add(-time.Hour*12).UnixNano(),
			time.Now().Add(-time.Hour*6).UnixNano(),
		)),
	}
	listResp := &shared.ListArchivedWorkflowExecutionsResponse{}
	var executions []*shared.WorkflowExecutionInfo
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		listResp, err = client.ListArchivedWorkflow(ctx, listReq)
		if err != nil {
			return err
		}

		if len(listResp.Executions) != 0 {
			executions = append(executions, listResp.Executions...)
		}

		if listResp.NextPageToken == nil {
			break
		}

		listReq.NextPageToken = listResp.NextPageToken
	}

	if len(listResp.Executions) == 0 {
		return errors.New("list archived workflow returned empty result")
	}

	return nil
}

func getURIScheme(URI string) string {
	if idx := strings.Index(URI, "://"); idx != -1 {
		return URI[:idx]
	}
	return ""
}

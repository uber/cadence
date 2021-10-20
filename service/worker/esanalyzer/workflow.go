// Copyright (c) 2017 Uber Technologies, Inc.
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

package esanalyzer

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

const (
	analyzerContextKey       = "analyzerContext"
	taskListName             = "cadence-sys-es-analyzer"
	wfTypeName               = "cadence-sys-es-analyzer-workflow"
	getWorkflowTypesActivity = "cadence-sys-es-analyzer-get-workflow-types"
)

var (
	retryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: time.Hour,
	}

	activityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: wfTypeName})
	activity.RegisterWithOptions(GetWorkflowTypes, activity.RegisterOptions{Name: getWorkflowTypesActivity})
}

// Workflow queries ElasticSearch to detect issues and mitigates them
func Workflow(ctx workflow.Context) error {
	opt := workflow.WithActivityOptions(ctx, activityOptions)
	_ = workflow.ExecuteActivity(opt, getWorkflowTypesActivity).Get(ctx, nil)
	return nil
}

// GetWorkflowTypes is activity to get workflow type list from ElasticSearch
func GetWorkflowTypes(ctx context.Context) error {
	/*
		analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
		analyzer.esClient.SearchByQuery(ctx, &es.SearchByQueryRequest{
			Index: analyzer.visibilityIndexName,
			Query: queryDSL,
			// NextPageToken:   request.NextPageToken,
			PageSize:        1000,
			Filter:          nil,
			MaxResultWindow: 1000,
		})
	*/
	return nil
}

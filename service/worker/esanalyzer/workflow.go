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
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/types"
)

type (
	contextKey string
)

const (
	analyzerContextKey contextKey = "analyzerContext"
	startUpDelay                  = time.Second * 10

	// workflow constants
	esAnalyzerWFID           = "cadence-sys-tl-esanalyzer"
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

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           esAnalyzerWFID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 5 * 24 * time.Hour,
		CronSchedule:                 "*/1 * * * *", // TODO: change the schedule
	}
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: wfTypeName})
	activity.RegisterWithOptions(GetWorkflowTypes, activity.RegisterOptions{Name: getWorkflowTypesActivity})
}

// Workflow queries ElasticSearch to detect issues and mitigates them
func Workflow(ctx workflow.Context) error {
	fmt.Printf("---------- Workflow execution starts: %v \n", activityOptions)
	opt := workflow.WithActivityOptions(ctx, activityOptions)
	var wfTypes []string
	err := workflow.ExecuteActivity(opt, getWorkflowTypesActivity).Get(ctx, &wfTypes)
	if err != nil {
		return err
	}

	fmt.Printf("---------- Workflow types: %v \n", wfTypes)
	return nil
}

// GetWorkflowTypes is activity to get workflow type list from ElasticSearch
func GetWorkflowTypes(ctx context.Context) (*[]string, error) {
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).UnixNano()
	// Get up to 1000 workflow types having at least 1 workflow in last 30 days
	// TODO: make #workflows and lastNDays configurable
	query := fmt.Sprintf(`
		{
			"query": {
					"bool": {
							"must": [
									{
											"range" : {
													"StartTime" : {
															"gte" : "%d"
													}
											}
									},
									{
											"exists": {
													"field": "CloseTime"
											}
									}
							]
					}
			},
			"size": 0,
			"aggs" : {
					"wfTypes" : {
							"terms" : { "field" : "WorkflowType", "size": 1000 }
					}
			}
		}
	`, thirtyDaysAgo)
	response, err := analyzer.esClient.SearchRaw(ctx, analyzer.visibilityIndexName, query)
	if err != nil {
		return nil, err
	}
	agg, foundAggregation := response.Aggregations["wfTypes"]
	if !foundAggregation {
		return nil, types.InternalServiceError{
			Message: "ElasticSearch aggeration failed",
		}
	}
	buckets := agg.Buckets
	if buckets == nil {
		return nil, types.InternalServiceError{
			Message: "ElasticSearch Error didn't return any workflow types",
		}
	}
	wfTypes := []string{}
	for _, bucket := range *buckets {
		wfTypes = append(wfTypes, bucket.Key)
	}
	return &wfTypes, nil
}

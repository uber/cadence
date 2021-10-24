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
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	contextKey string

	Duration struct {
		AvgExecTime float64 `json:"value"`
	}

	WorkflowTypeInfo struct {
		Name         string   `json:"key"`
		NumWorfklows int64    `json:"doc_count"`
		Duration     Duration `json:"duration"`
	}
)

const (
	analyzerContextKey contextKey = "analyzerContext"
	startUpDelay                  = time.Second * 10

	// workflow constants
	esAnalyzerWFID                = "cadence-sys-tl-esanalyzer"
	taskListName                  = "cadence-sys-es-analyzer"
	wfTypeName                    = "cadence-sys-es-analyzer-workflow"
	getWorkflowTypesActivity      = "cadence-sys-es-analyzer-get-workflow-types"
	findStuckWorkflowsActivity    = "cadence-sys-es-analyzer-find-stuck-workflows"
	refreshStuckWorkflowsActivity = "cadence-sys-es-analyzer-refresh-stuck-workflows"
)

var (
	startDateTime = time.Now().AddDate(0, 0, -30).UnixNano()

	retryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: time.Hour,
	}

	getWorkflowTypesOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}
	findStuckWorkflowsOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    3 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}
	refreshStuckWorkflowsOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
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
	activity.RegisterWithOptions(FindStuckWorkflows, activity.RegisterOptions{Name: findStuckWorkflowsActivity})
	activity.RegisterWithOptions(
		RefreshStuckWorkflowsFromSameWorkflowType,
		activity.RegisterOptions{Name: refreshStuckWorkflowsActivity},
	)
}

// Workflow queries ElasticSearch to detect issues and mitigates them
func Workflow(ctx workflow.Context) error {
	// list of workflows with avg workflow duration
	var wfTypes []WorkflowTypeInfo
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getWorkflowTypesOptions),
		getWorkflowTypesActivity,
	).Get(ctx, &wfTypes)
	if err != nil {
		return err
	}

	for _, info := range wfTypes {
		// not enough workflows to get avg time, consider making it configurable
		if info.NumWorfklows < 100 {
			continue
		}
		var stuckWorkflows []*persistence.InternalVisibilityWorkflowExecutionInfo
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, findStuckWorkflowsOptions),
			findStuckWorkflowsActivity,
			info,
		).Get(ctx, &stuckWorkflows)
		if err != nil {
			return err
		}
		if len(stuckWorkflows) == 0 {
			continue
		}

		err = workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, refreshStuckWorkflowsOptions),
			refreshStuckWorkflowsActivity,
			stuckWorkflows,
		).Get(ctx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// RefreshStuckWorkflowsFromSameWorkflowType is activity to refresh stuck workflows from the same domain
func RefreshStuckWorkflowsFromSameWorkflowType(
	ctx context.Context,
	workflows []*persistence.InternalVisibilityWorkflowExecutionInfo,
) error {
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	domainID := workflows[0].DomainID
	domainEntry, err := analyzer.domainCache.GetDomainByID(domainID)
	if err != nil {
		analyzer.logger.Error("Failed to get domain entry",
			tag.WorkflowDomainID(domainID),
		)
		return err
	}
	domainName := domainEntry.GetInfo().Name
	clusterName := domainEntry.GetReplicationConfig().ActiveClusterName

	adminClient := analyzer.clientBean.GetRemoteAdminClient(clusterName)
	for _, workflow := range workflows {
		if workflow.DomainID != domainID {
			return types.InternalServiceError{
				Message: fmt.Sprintf(
					"Inconsistent worklow. Expected domainID: %v, actual: %v",
					domainID,
					workflow.DomainID),
			}
		}

		err = adminClient.RefreshWorkflowTasks(ctx, &types.RefreshWorkflowTasksRequest{
			Domain: domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: workflow.WorkflowID,
				RunID:      workflow.RunID,
			},
		})

		if err != nil {
			// Errors might happen if the workflow is already closed. Instead of failing the workflow
			// log the error and continue
			analyzer.logger.Error("Failed to refresh stuck workflow",
				tag.WorkflowDomainName(domainName),
				tag.WorkflowID(workflow.WorkflowID),
				tag.WorkflowRunID(workflow.RunID),
			)
		} else {
			analyzer.logger.Info("Refreshed stuck workflow",
				tag.WorkflowDomainName(domainName),
				tag.WorkflowID(workflow.WorkflowID),
				tag.WorkflowRunID(workflow.RunID),
			)
		}
	}

	return nil
}

// FindStuckWorkflows is activity to find open workflows that are live significantly longer than average
func FindStuckWorkflows(ctx context.Context, info WorkflowTypeInfo) (*[]*persistence.InternalVisibilityWorkflowExecutionInfo, error) {
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	// allow some buffer time to any workflow; consider making this configurable
	maxEndTimeAllowed := time.Now().Add(
		-time.Second * time.Duration((int64(info.Duration.AvgExecTime) * 3)),
	).UnixNano()
	endTime := time.Now().Add(-time.Minute * time.Duration(30)).UnixNano()

	if endTime > maxEndTimeAllowed {
		endTime = maxEndTimeAllowed
	}

	query := fmt.Sprintf(`
		{
			"query": {
					"bool": {
							"must": [
									{
											"range" : {
													"StartTime" : {
															"gte" : "%d",
															"lte" : "%d"
													}
											}
									},
									{
											"match" : {
													"WorkflowType" : "%s"
											}
									}
							],
						  "must_not": {
							  "exists": {
								 	"field": "CloseTime"
							  }
						  }
					 }
			 }
		}
		`, startDateTime, endTime, info.Name)

	response, err := analyzer.esClient.SearchRaw(ctx, analyzer.visibilityIndexName, query)
	if err != nil {
		analyzer.logger.Error("Failed to query ElasticSearch for stuck workflows",
			tag.VisibilityQuery(query),
		)
		return nil, err
	}

	return &response.Hits.Hits, nil
}

// GetWorkflowTypes is activity to get workflow type list from ElasticSearch
func GetWorkflowTypes(ctx context.Context) (*[]WorkflowTypeInfo, error) {
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	aggregationKey := "wfTypes"
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
					"%s" : {
							"terms" : { "field" : "WorkflowType", "size": 10 },
							"aggs": {
									"duration" : {
										"avg" : {
											"script" : "(doc['CloseTime'].value - doc['StartTime'].value) / 1000000000"
										}
									}
							}
					}
			}
		}
	`, startDateTime, aggregationKey)

	response, err := analyzer.esClient.SearchRaw(ctx, analyzer.visibilityIndexName, query)
	if err != nil {
		analyzer.logger.Error("Failed to query ElasticSearch to find workflow type info",
			tag.VisibilityQuery(query),
		)
		return nil, err
	}
	agg, foundAggregation := response.Aggregations[aggregationKey]
	if !foundAggregation {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch error: aggeration '%v' failed", aggregationKey),
		}
	}

	var wfTypes struct {
		Buckets []WorkflowTypeInfo `json:"buckets"`
	}
	err = json.Unmarshal(*agg, &wfTypes)
	if err != nil {
		return nil, types.InternalServiceError{
			Message: "ElasticSearch error parsing aggeration",
		}
	}

	// This log is supposed to be fired at max once an hour; it's not invasive and can help
	// get some workflow statistics. Size can be quite big though; not sure what the limit is.
	analyzer.logger.Info(fmt.Sprintf("WorkflowType stats: %#v", wfTypes.Buckets))
	return &wfTypes.Buckets, nil
}

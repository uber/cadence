// Copyright (c) 2021 Uber Technologies, Inc.
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
	"strings"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	contextKey string

	Duration struct {
		AvgExecTimeNanoseconds float64 `json:"value"`
	}

	WorkflowTypeInfo struct {
		Name         string   `json:"key"`
		NumWorkflows int64    `json:"doc_count"`
		Duration     Duration `json:"duration"`
	}

	WorkflowInfo struct {
		DomainID   string `json:"DomainID"`
		WorkflowID string `json:"WorkflowID"`
		RunID      string `json:"RunID"`
	}
)

const (
	analyzerContextKey contextKey = "analyzerContext"

	// workflow constants
	esAnalyzerWFID                = "cadence-sys-tl-esanalyzer"
	taskListName                  = "cadence-sys-es-analyzer"
	esanalyzerWFTypeName          = "cadence-sys-es-analyzer-workflow"
	getWorkflowTypesActivity      = "cadence-sys-es-analyzer-get-workflow-types"
	findStuckWorkflowsActivity    = "cadence-sys-es-analyzer-find-stuck-workflows"
	refreshStuckWorkflowsActivity = "cadence-sys-es-analyzer-refresh-stuck-workflows"
)

var (
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
		ExecutionStartToCloseTimeout: 24 * time.Hour,
		CronSchedule:                 "0 * * * *", // "At minute 0" => every hour
	}
)

func init() {
	workflow.RegisterWithOptions(workflowFunc, workflow.RegisterOptions{Name: esanalyzerWFTypeName})
	activity.RegisterWithOptions(getWorkflowTypes, activity.RegisterOptions{Name: getWorkflowTypesActivity})
	activity.RegisterWithOptions(findStuckWorkflows, activity.RegisterOptions{Name: findStuckWorkflowsActivity})
	activity.RegisterWithOptions(
		refreshStuckWorkflowsFromSameWorkflowType,
		activity.RegisterOptions{Name: refreshStuckWorkflowsActivity},
	)
}

// workflowFunc queries ElasticSearch to detect issues and mitigates them
func workflowFunc(ctx workflow.Context) error {
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
		var stuckWorkflows []WorkflowInfo
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

// refreshStuckWorkflowsFromSameWorkflowType is activity to refresh stuck workflows from the same domain
func refreshStuckWorkflowsFromSameWorkflowType(
	ctx context.Context,
	workflows []WorkflowInfo,
) error {
	logger := activity.GetLogger(ctx)
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	domainID := workflows[0].DomainID
	domainEntry, err := analyzer.domainCache.GetDomainByID(domainID)
	if err != nil {
		logger.Error("Failed to get domain entry",
			zap.String("error", fmt.Sprintf("%v", err)),
			zap.String("DomainID", domainID))
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
			logger.Error("Failed to refresh stuck workflow",
				zap.String("error", fmt.Sprintf("%v", err)),
				zap.String("domainName", domainName),
				zap.String("workflowID", workflow.WorkflowID),
				zap.String("runID", workflow.RunID),
			)
			analyzer.metricsClient.IncCounter(metrics.ESAnalyzerScope, metrics.ESAnalyzerNumStuckWorkflowsFailedToRefresh)
		} else {
			logger.Info("Refreshed stuck workflow",
				zap.String("domainName", domainName),
				zap.String("workflowID", workflow.WorkflowID),
				zap.String("runID", workflow.RunID),
			)
			analyzer.metricsClient.IncCounter(metrics.ESAnalyzerScope, metrics.ESAnalyzerNumStuckWorkflowsRefreshed)
		}
	}

	return nil
}

func getFindStuckWorkflowsQuery(
	startDateTime int64,
	endTime int64,
	workflowType string,
	maxNumWorkflows int,
) (string, error) {
	wfTypeMarshaled, err := json.Marshal(workflowType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
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
      },
      "size": %d
    }
    `, startDateTime, endTime, wfTypeMarshaled, maxNumWorkflows), nil
}

// findStuckWorkflows is activity to find open workflows that are live significantly longer than average
func findStuckWorkflows(ctx context.Context, info WorkflowTypeInfo) ([]WorkflowInfo, error) {
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	logger := activity.GetLogger(ctx)

	minNumWorkflowsNeeded := int64(analyzer.config.ESAnalyzerMinNumWorkflowsForAvg(info.Name))
	if len(analyzer.config.ESAnalyzerLimitToTypes()) > 0 {
		logger.Info("Skipping minimum workflow count validation since workflow types were passed from config")
	} else if info.NumWorkflows < minNumWorkflowsNeeded {
		logger.Warn(fmt.Sprintf(
			"Skipping workflow type '%s' because it doesn't have enough(%d) workflows to avg",
			info.Name,
			minNumWorkflowsNeeded,
		))
		return nil, nil
	}

	startDateTime := time.Now().Add(-analyzer.config.ESAnalyzerTimeWindow()).UnixNano()

	// allow some buffer time to any workflow
	maxEndTimeAllowed := time.Now().Add(-analyzer.config.ESAnalyzerBufferWaitTime(info.Name)).UnixNano()

	// if the workflow exec time takes longer than 3x avg time, we refresh
	endTime := time.Now().Add(
		-time.Nanosecond * time.Duration((int64(info.Duration.AvgExecTimeNanoseconds) * 3)),
	).UnixNano()
	if endTime > maxEndTimeAllowed {
		endTime = maxEndTimeAllowed
	}

	maxNumWorkflows := analyzer.config.ESAnalyzerNumWorkflowsToRefresh(info.Name)
	query, err := getFindStuckWorkflowsQuery(startDateTime, endTime, info.Name, maxNumWorkflows)
	if err != nil {
		logger.Error("Failed to create ElasticSearch query for stuck workflows",
			zap.String("error", fmt.Sprintf("%v", err)),
			zap.String("startDateTime", fmt.Sprintf("%v", startDateTime)),
			zap.String("endTime", fmt.Sprintf("%v", endTime)),
			zap.String("workflowType", info.Name),
			zap.String("maxNumWorkflows", fmt.Sprintf("%v", maxNumWorkflows)),
		)
		return nil, err
	}
	response, err := analyzer.esClient.SearchRaw(ctx, analyzer.visibilityIndexName, query)
	if err != nil {
		logger.Error("Failed to query ElasticSearch for stuck workflows",
			zap.String("error", fmt.Sprintf("%v", err)),
			zap.String("VisibilityQuery", query),
		)
		return nil, err
	}

	// Return a simpler structure to reduce activity output size
	workflows := []WorkflowInfo{}
	if response.Hits.Hits != nil {
		for _, hit := range response.Hits.Hits {
			workflows = append(workflows, WorkflowInfo{
				DomainID:   hit.DomainID,
				WorkflowID: hit.WorkflowID,
				RunID:      hit.RunID,
			})
		}
	}

	if len(workflows) > 0 {
		analyzer.metricsClient.AddCounter(
			metrics.ESAnalyzerScope,
			metrics.ESAnalyzerNumStuckWorkflowsDiscovered,
			int64(len(workflows)))
	}

	return workflows, nil
}

func getWorkflowTypesQuery(startDateTime int64, aggregationKey string) string {
	return fmt.Sprintf(`
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
                      "script" : "(doc['CloseTime'].value - doc['StartTime'].value)"
                    }
                  }
              }
          }
      }
    }
	`, startDateTime, aggregationKey)
}

// getWorkflowTypes is activity to get workflow type list from ElasticSearch
func getWorkflowTypes(ctx context.Context) ([]WorkflowTypeInfo, error) {
	logger := activity.GetLogger(ctx)
	analyzer := ctx.Value(analyzerContextKey).(*Analyzer)
	limitToTypes := analyzer.config.ESAnalyzerLimitToTypes()
	if len(limitToTypes) > 0 {
		results := []WorkflowTypeInfo{}
		workflowTypes := strings.Split(limitToTypes, ",")
		for _, wfType := range workflowTypes {
			results = append(results, WorkflowTypeInfo{Name: wfType})
		}
		return results, nil
	}

	aggregationKey := "wfTypes"
	startDateTime := time.Now().Add(-analyzer.config.ESAnalyzerTimeWindow()).UnixNano()
	query := getWorkflowTypesQuery(startDateTime, aggregationKey)
	response, err := analyzer.esClient.SearchRaw(ctx, analyzer.visibilityIndexName, query)
	if err != nil {
		logger.Error("Failed to query ElasticSearch to find workflow type info",
			zap.String("error", fmt.Sprintf("%v", err)),
			zap.String("VisibilityQuery", query),
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
	err = json.Unmarshal(agg, &wfTypes)
	if err != nil {
		return nil, types.InternalServiceError{
			Message: "ElasticSearch error parsing aggeration",
		}
	}

	// This log is supposed to be fired at max once an hour; it's not invasive and can help
	// get some workflow statistics. Size can be quite big though; not sure what the limit is.
	logger.Info(fmt.Sprintf("WorkflowType stats: %#v", wfTypes.Buckets))
	return wfTypes.Buckets, nil
}

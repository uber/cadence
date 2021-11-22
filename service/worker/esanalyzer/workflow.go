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

const (
	domainsAggKey = "domains"
	wfTypesAggKey = "wfTypes"

	// workflow constants
	esAnalyzerWFID                   = "cadence-sys-tl-esanalyzer"
	taskListName                     = "cadence-sys-es-analyzer"
	esanalyzerWFTypeName             = "cadence-sys-es-analyzer-workflow"
	getWorkflowTypesActivity         = "cadence-sys-es-analyzer-get-workflow-types"
	findStuckWorkflowsActivity       = "cadence-sys-es-analyzer-find-stuck-workflows"
	refreshStuckWorkflowsActivity    = "cadence-sys-es-analyzer-refresh-stuck-workflows"
	findLongRunningWorkflowsActivity = "cadence-sys-es-analyzer-find-long-running-workflows"
)

type (
	Workflow struct {
		analyzer *Analyzer
	}
	Duration struct {
		AvgExecTimeNanoseconds float64 `json:"value"`
	}

	WorkflowTypeInfo struct {
		DomainID     string   // this won't come from ES result json
		Name         string   `json:"key"`
		NumWorkflows int64    `json:"doc_count"`
		Duration     Duration `json:"duration"`
	}

	WorkflowTypeInfoContainer struct {
		WorkflowTypes []WorkflowTypeInfo `json:"buckets"`
	}

	DomainInfo struct {
		DomainID        string                    `json:"key"`
		NumWorkflows    int64                     `json:"doc_count"`
		WFTypeContainer WorkflowTypeInfoContainer `json:"wfTypes"`
	}

	WorkflowInfo struct {
		DomainID   string `json:"DomainID"`
		WorkflowID string `json:"WorkflowID"`
		RunID      string `json:"RunID"`
	}
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
	findLongRunningWorkflowsOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    15 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           esAnalyzerWFID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 24 * time.Hour,
		CronSchedule:                 "0 * * * *", // "At minute 0" => every hour
	}
)

func initWorkflow(a *Analyzer) {
	w := Workflow{analyzer: a}
	workflow.RegisterWithOptions(w.workflowFunc, workflow.RegisterOptions{Name: esanalyzerWFTypeName})
	activity.RegisterWithOptions(w.getWorkflowTypes, activity.RegisterOptions{Name: getWorkflowTypesActivity})
	activity.RegisterWithOptions(w.findStuckWorkflows, activity.RegisterOptions{Name: findStuckWorkflowsActivity})
	activity.RegisterWithOptions(
		w.refreshStuckWorkflowsFromSameWorkflowType,
		activity.RegisterOptions{Name: refreshStuckWorkflowsActivity})
	activity.RegisterWithOptions(
		w.findLongRunningWorkflows,
		activity.RegisterOptions{Name: findLongRunningWorkflowsActivity})
}

// workflowFunc queries ElasticSearch to detect issues and mitigates them
func (w *Workflow) workflowFunc(ctx workflow.Context) error {
	if w.analyzer.config.ESAnalyzerPause() {
		logger := workflow.GetLogger(ctx)
		logger.Info("Skipping ESAnalyzer execution cycle since it was paused")
		return nil
	}

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

	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, findLongRunningWorkflowsOptions),
		findLongRunningWorkflowsActivity,
	).Get(ctx, nil)

	return err
}

func getLongRunningWorkflowsQuery(
	maxWorkflowStartTime int64,
	domainID string,
	workflowType string,
) (string, error) {
	wfTypeMarshaled, err := json.Marshal(workflowType)
	if err != nil {
		return "", err
	}
	// No need to marshal domainID: it comes from domainEntry and its type is uuid
	return fmt.Sprintf(`
    {
      "query": {
          "bool": {
              "must": [
                  {
                      "range" : {
                          "StartTime" : {
                              "lte" : "%d"
                          }
                      }
                  },
                  {
                      "match" : {
                          "DomainID" : "%s"
                      }
                  },
                  {
                      "match" : {
                          "WorkflowType" : %s
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
      "size": 0
    }
    `, maxWorkflowStartTime, domainID, string(wfTypeMarshaled)), nil
}

// findLongRunningWorkflows is activity to find long running workflows
func (w *Workflow) findLongRunningWorkflows(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	workflowThresholds := w.analyzer.config.ESAnalyzerWorkflowDurationWarnThresholds()
	if len(workflowThresholds) == 0 {
		logger.Info("No workflow execution time thresholds defined")
		return nil
	}

	var entries map[string]string
	err := json.Unmarshal([]byte(workflowThresholds), &entries)
	if err != nil {
		return err
	}

	domainNameToID := map[string]string{}
	for domainWFTypePair, threshold := range entries {
		index := strings.Index(domainWFTypePair, "/")
		// -1 no delimiter, 0 means the entry starts with /
		if index < 1 || len(domainWFTypePair) <= (index+1) {
			return types.InternalServiceError{
				Message: fmt.Sprintf("Bad Workflow type entry %q", domainWFTypePair),
			}
		}
		domainName := domainWFTypePair[:index]
		wfType := domainWFTypePair[index+1:]
		var domainID string
		if _, ok := domainNameToID[domainName]; !ok {
			domainEntry, err := w.analyzer.domainCache.GetDomain(domainName)
			if err != nil {
				logger.Error("Failed to get domain entry",
					zap.Error(err),
					zap.String("DomainName", domainName))
				return err
			}
			domainID = domainEntry.GetInfo().ID
			domainNameToID[domainName] = domainID
		}

		durVal, err := time.ParseDuration(threshold)
		if err != nil {
			return err
		}
		maxWorkflowStartTime := time.Now().Add(-durVal).UnixNano()

		query, err := getLongRunningWorkflowsQuery(maxWorkflowStartTime, domainID, wfType)
		if err != nil {
			return err
		}
		response, err := w.analyzer.esClient.SearchRaw(ctx, w.analyzer.visibilityIndexName, query)
		if err != nil {
			logger.Error("Failed to query ElasticSearch for stuck workflows",
				zap.Error(err),
				zap.String("VisibilityQuery", query),
			)
			return err
		}

		if response.Hits.TotalHits > 0 {
			logger.Warn("Slow running workflows detected",
				zap.String("DomainName", domainName),
				zap.String("WorkflowType", wfType))
		}

		tagged := w.analyzer.scopedMetricClient.Tagged(
			metrics.DomainTag(domainName),
			metrics.WorkflowTypeTag(wfType),
		)
		tagged.AddCounter(
			metrics.ESAnalyzerNumLongRunningWorkflows,
			response.Hits.TotalHits)
	}

	return nil
}

// refreshStuckWorkflowsFromSameWorkflowType is activity to refresh stuck workflows from the same domain
func (w *Workflow) refreshStuckWorkflowsFromSameWorkflowType(
	ctx context.Context,
	workflows []WorkflowInfo,
) error {
	logger := activity.GetLogger(ctx)
	domainID := workflows[0].DomainID
	domainEntry, err := w.analyzer.domainCache.GetDomainByID(domainID)
	if err != nil {
		logger.Error("Failed to get domain entry", zap.Error(err), zap.String("DomainID", domainID))
		return err
	}
	domainName := domainEntry.GetInfo().Name
	clusterName := domainEntry.GetReplicationConfig().ActiveClusterName

	adminClient := w.analyzer.clientBean.GetRemoteAdminClient(clusterName)
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
				zap.Error(err),
				zap.String("domainName", domainName),
				zap.String("workflowID", workflow.WorkflowID),
				zap.String("runID", workflow.RunID),
			)
			w.analyzer.scopedMetricClient.IncCounter(metrics.ESAnalyzerNumStuckWorkflowsFailedToRefresh)
		} else {
			logger.Info("Refreshed stuck workflow",
				zap.String("domainName", domainName),
				zap.String("workflowID", workflow.WorkflowID),
				zap.String("runID", workflow.RunID),
			)
			w.analyzer.scopedMetricClient.IncCounter(metrics.ESAnalyzerNumStuckWorkflowsRefreshed)
		}
	}

	return nil
}

func getFindStuckWorkflowsQuery(
	startDateTime int64,
	endTime int64,
	domainID string,
	workflowType string,
	maxNumWorkflows int,
) (string, error) {
	wfTypeMarshaled, err := json.Marshal(workflowType)
	if err != nil {
		return "", err
	}
	// No need to marshal domainID: it comes from domainEntry and its type is uuid
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
                          "DomainID" : "%s"
                      }
                  },
                  {
                      "match" : {
                          "WorkflowType" : %s
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
    `, startDateTime, endTime, domainID, string(wfTypeMarshaled), maxNumWorkflows), nil
}

// findStuckWorkflows is activity to find open workflows that are live significantly longer than average
func (w *Workflow) findStuckWorkflows(ctx context.Context, info WorkflowTypeInfo) ([]WorkflowInfo, error) {
	logger := activity.GetLogger(ctx)
	domainEntry, err := w.analyzer.domainCache.GetDomainByID(info.DomainID)
	if err != nil {
		logger.Error("Failed to get domain entry", zap.Error(err), zap.String("DomainID", info.DomainID))
		return nil, err
	}
	domainName := domainEntry.GetInfo().Name

	minNumWorkflowsNeeded := int64(w.analyzer.config.ESAnalyzerMinNumWorkflowsForAvg(domainName, info.Name))
	if len(w.analyzer.config.ESAnalyzerLimitToTypes()) > 0 || len(w.analyzer.config.ESAnalyzerLimitToDomains()) > 0 {
		logger.Info("Skipping minimum workflow count validation since workflow types were passed from config")
	} else if info.NumWorkflows < minNumWorkflowsNeeded {
		logger.Warn(fmt.Sprintf(
			"Skipping workflow type '%s' because it doesn't have enough(%d) workflows to avg",
			info.Name,
			minNumWorkflowsNeeded,
		))
		return nil, nil
	}

	startDateTime := time.Now().Add(-w.analyzer.config.ESAnalyzerTimeWindow()).UnixNano()

	// allow some buffer time to any workflow
	maxEndTimeAllowed := time.Now().Add(
		-w.analyzer.config.ESAnalyzerBufferWaitTime(domainName, info.Name),
	).UnixNano()

	// if the workflow exec time takes longer than 3x avg time, we refresh
	endTime := time.Now().Add(
		-time.Duration((int64(info.Duration.AvgExecTimeNanoseconds) * 3)),
	).UnixNano()
	if endTime > maxEndTimeAllowed {
		endTime = maxEndTimeAllowed
	}

	maxNumWorkflows := w.analyzer.config.ESAnalyzerNumWorkflowsToRefresh(domainName, info.Name)
	query, err := getFindStuckWorkflowsQuery(startDateTime, endTime, info.DomainID, info.Name, maxNumWorkflows)
	if err != nil {
		logger.Error("Failed to create ElasticSearch query for stuck workflows",
			zap.Error(err),
			zap.Int64("startDateTime", startDateTime),
			zap.Int64("endTime", endTime),
			zap.String("workflowType", info.Name),
			zap.Int("maxNumWorkflows", maxNumWorkflows),
		)
		return nil, err
	}
	response, err := w.analyzer.esClient.SearchRaw(ctx, w.analyzer.visibilityIndexName, query)
	if err != nil {
		logger.Error("Failed to query ElasticSearch for stuck workflows",
			zap.Error(err),
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
		w.analyzer.scopedMetricClient.AddCounter(
			metrics.ESAnalyzerNumStuckWorkflowsDiscovered,
			int64(len(workflows)))
	}

	return workflows, nil
}

func (w *Workflow) getDomainsLimitQuery() (string, error) {
	limitToDomains := w.analyzer.config.ESAnalyzerLimitToDomains()

	domainsLimitQuery := ""
	if len(limitToDomains) > 0 {
		var domainNames []string
		err := json.Unmarshal([]byte(limitToDomains), &domainNames)
		if err != nil {
			return "", err
		}
		if len(domainNames) > 0 {
			domainIDs := []string{}
			for _, domainName := range domainNames {
				domainEntry, err := w.analyzer.domainCache.GetDomain(domainName)
				if err != nil {
					return "", err
				}
				domainIDs = append(domainIDs, domainEntry.GetInfo().ID)
			}

			marshaledDomains, err := json.Marshal(domainIDs)
			if err != nil {
				return "", err
			}
			domainsLimitQuery = fmt.Sprintf(`,
				{
						"terms" : {
								"DomainID" : %s
						}
				}
			`, string(marshaledDomains))
		}
	}
	return domainsLimitQuery, nil
}

func (w *Workflow) getWorkflowTypesQuery() (string, error) {
	domainsLimitQuery, err := w.getDomainsLimitQuery()
	if err != nil {
		return "", nil
	}
	startDateTime := time.Now().Add(-w.analyzer.config.ESAnalyzerTimeWindow()).UnixNano()
	maxNumDomains := w.analyzer.config.ESAnalyzerMaxNumDomains()
	maxNumWorkflowTypes := w.analyzer.config.ESAnalyzerMaxNumWorkflowTypes()

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
						%s
          ]
        }
      },
      "size": 0,
      "aggs" : {
        "%s" : {
          "terms" : { "field" : "DomainID", "size": %d },
          "aggs": {
            "%s" : {
              "terms" : { "field" : "WorkflowType", "size": %d },
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
      }
    }
	`, startDateTime, domainsLimitQuery, domainsAggKey, maxNumDomains, wfTypesAggKey, maxNumWorkflowTypes), nil
}

func (w *Workflow) getWorkflowTypesFromDynamicConfig(
	ctx context.Context,
	config string,
	logger *zap.Logger,
) ([]WorkflowTypeInfo, error) {
	results := []WorkflowTypeInfo{}
	var entries []string
	err := json.Unmarshal([]byte(config), &entries)
	if err != nil {
		return nil, err
	}

	domainNameToID := map[string]string{}

	for _, domainWFTypePair := range entries {
		index := strings.Index(domainWFTypePair, "/")
		// -1 no delimiter, 0 means the entry starts with /
		if index < 1 || len(domainWFTypePair) <= (index+1) {
			return nil, types.InternalServiceError{
				Message: fmt.Sprintf("Bad Workflow type entry %q", domainWFTypePair),
			}
		}
		domainName := domainWFTypePair[:index]
		wfType := domainWFTypePair[index+1:]
		if _, ok := domainNameToID[domainName]; !ok {
			domainEntry, err := w.analyzer.domainCache.GetDomain(domainName)
			if err != nil {
				logger.Error("Failed to get domain entry",
					zap.Error(err),
					zap.String("DomainName", domainName))
				return nil, err
			}
			domainNameToID[domainName] = domainEntry.GetInfo().ID
		}

		results = append(results, WorkflowTypeInfo{
			DomainID: domainNameToID[domainName],
			Name:     wfType,
		})
	}

	return results, nil

}

func normalizeDomainInfos(infos []DomainInfo) []WorkflowTypeInfo {
	results := []WorkflowTypeInfo{}
	for _, domainInfo := range infos {
		for _, wfType := range domainInfo.WFTypeContainer.WorkflowTypes {
			results = append(results, WorkflowTypeInfo{
				DomainID: domainInfo.DomainID,
				Name:     wfType.Name,
			})
		}
	}
	return results
}

// getWorkflowTypes is activity to get workflow type list from ElasticSearch
func (w *Workflow) getWorkflowTypes(ctx context.Context) ([]WorkflowTypeInfo, error) {
	logger := activity.GetLogger(ctx)

	limitToTypes := w.analyzer.config.ESAnalyzerLimitToTypes()
	if len(limitToTypes) > 0 {
		return w.getWorkflowTypesFromDynamicConfig(ctx, limitToTypes, logger)
	}

	query, err := w.getWorkflowTypesQuery()
	if err != nil {
		return nil, err
	}

	response, err := w.analyzer.esClient.SearchRaw(ctx, w.analyzer.visibilityIndexName, query)
	if err != nil {
		logger.Error("Failed to query ElasticSearch to find workflow type info",
			zap.Error(err),
			zap.String("VisibilityQuery", query),
		)
		return nil, err
	}
	agg, foundAggregation := response.Aggregations[domainsAggKey]
	if !foundAggregation {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch error: aggeration failed. Query: %v", query),
		}
	}

	var domains struct {
		Buckets []DomainInfo `json:"buckets"`
	}
	err = json.Unmarshal(agg, &domains)
	if err != nil {
		return nil, types.InternalServiceError{
			Message: "ElasticSearch error parsing aggeration",
		}
	}

	// This log is supposed to be fired at max once an hour; it's not invasive and can help
	// get some workflow statistics. Size can be quite big though; not sure what the limit is.
	logger.Info(fmt.Sprintf("WorkflowType stats: %#v", domains.Buckets))

	return normalizeDomainInfos(domains.Buckets), nil
}

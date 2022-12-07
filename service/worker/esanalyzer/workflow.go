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
	"github.com/uber/cadence/common/metrics"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"time"

	"go.uber.org/cadence"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
)

const (
	domainsAggKey = "domains"
	wfTypesAggKey = "wfTypes"

	// workflow constants
	esAnalyzerWFID                     = "cadence-sys-tl-esanalyzer"
	taskListName                       = "cadence-sys-es-analyzer"
	esanalyzerWFTypeName               = "cadence-sys-es-analyzer-workflow"
	emitWorkflowVersionMetricsActivity = "cadence-sys-es-analyzer-emit-workflow-version-metrics"
)

type (
	Workflow struct {
		analyzer *Analyzer
	}

	DomainWorkflowVersionCount struct {
		WorkflowVersions []WorkflowVersionCount `json:"buckets"`
	}
	WorkflowVersionCount struct {
		WorkflowVersion string `json:"key"`
		NumWorkflows    int64  `json:"doc_count"`
	}
)

var (
	retryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: time.Hour,
	}

	getWorkflowMetricsOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           esAnalyzerWFID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 1 * time.Hour,
		CronSchedule:                 "*/5 * * * *", // "At minute 0" => every hour
	}
)

func initWorkflow(a *Analyzer) {
	w := Workflow{analyzer: a}
	workflow.RegisterWithOptions(w.workflowFunc, workflow.RegisterOptions{Name: esanalyzerWFTypeName})
	activity.RegisterWithOptions(
		w.emitWorkflowVersionMetrics,
		activity.RegisterOptions{Name: emitWorkflowVersionMetricsActivity},
	)
}

// workflowFunc queries ElasticSearch to detect issues and mitigates them
func (w *Workflow) workflowFunc(ctx workflow.Context) error {
	if w.analyzer.config.ESAnalyzerPause() {
		logger := workflow.GetLogger(ctx)
		logger.Info("Skipping ESAnalyzer execution cycle since it was paused")
		return nil
	}
	var err error
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, getWorkflowMetricsOptions),
		emitWorkflowVersionMetricsActivity,
	).Get(ctx, &err)
	if err != nil {
		return err
	}
	return nil
}

func (w *Workflow) getWorkflowVersionQuery(domainName string) (string, error) {
	domain, err := w.analyzer.domainCache.GetDomain(domainName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
    {
      "size": 0,
      "query": {
        "bool": {
          "must": [
            {
              "exists": {
                "field": "Attr.CadenceChangeVersion"
              }
            },
            {
                "match": {
                    "DomainID": "%s"
                }
            }
          ],
          "must_not":{
              "exists":{
                  "field": "CloseTime"
              }
          }
        }
      },
        "aggs": {
            "WorkflowVersions": {
              "terms": {
                "field": "Attr.CadenceChangeVersion"
              }
            }
         }
    }
    `, domain.GetInfo().ID), nil
}

// emitWorkflowVersionMetrics is an activity that emits the running WF versions of a domain
func (w *Workflow) emitWorkflowVersionMetrics(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	var workflowMetricDomainNames []string
	workflowMetricDomains := w.analyzer.config.ESAnalyzerWorkflowVersionDomains()
	logger.Info("Domains to get metrics on", zap.String("DomainsToGetMetrics", workflowMetricDomains))
	if len(workflowMetricDomains) > 0 {
		err := json.Unmarshal([]byte(workflowMetricDomains), &workflowMetricDomainNames)
		if err != nil {
			return err
		}
		for _, domainName := range workflowMetricDomainNames {
			wfVersionEsQuery, err := w.getWorkflowVersionQuery(domainName)
			if err != nil {
				logger.Error("Failed to get ElasticSearch query to find workflow version Info",
					zap.Error(err),
					zap.String("DomainName", domainName),
				)
				return err
			}
			response, err := w.analyzer.esClient.SearchRaw(ctx, w.analyzer.visibilityIndexName, wfVersionEsQuery)
			if err != nil {
				logger.Error("Failed to query ElasticSearch to find workflow version Info",
					zap.Error(err),
					zap.String("VisibilityQuery", wfVersionEsQuery),
					zap.String("DomainName", domainName),
				)
				return err
			}
			agg, foundAggregation := response.Aggregations["WorkflowVersions"]
			logger.Info("raw response", zap.String("RawResponse", string(agg)))

			if !foundAggregation {
				logger.Error("ElasticSearch error: aggregation failed.",
					zap.Error(err),
					zap.String("Aggregation", string(agg)),
					zap.String("DomainName", domainName),
					zap.String("VisibilityQuery", wfVersionEsQuery),
				)
				return err
			}
			var domainWorkflowVersionCount DomainWorkflowVersionCount
			err = json.Unmarshal(agg, &domainWorkflowVersionCount)
			if err != nil {
				logger.Error("ElasticSearch error parsing aggregation.",
					zap.Error(err),
					zap.String("Aggregation", string(agg)),
					zap.String("DomainName", domainName),
					zap.String("VisibilityQuery", wfVersionEsQuery),
				)
				return err
			}
			for _, workflowVersion := range domainWorkflowVersionCount.WorkflowVersions {
				logger.Info("Workflow Version Emitting metric",
					zap.String("DomainName", domainName),
					zap.Int("VersionCount", int(workflowVersion.NumWorkflows)),
					zap.String("Workflow Version", workflowVersion.WorkflowVersion))
				w.analyzer.scopedMetricClient.Tagged(metrics.DomainTag(domainName),
					metrics.WorkflowVersionTag(workflowVersion.WorkflowVersion)).UpdateGauge(metrics.WorkflowVersionCount, float64(workflowVersion.NumWorkflows))
			}
		}
	}
	return nil
}

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
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	workflowTypesAggKey         = "wftypes"
	domainTag                   = "domain"
	workflowVersionTag          = "workflowVersion"
	workflowVersionCountMetrics = "workflow_version_count"
	workflowTypeTag             = "workflowType"

	// workflow constants
	esAnalyzerWFID                     = "cadence-sys-tl-esanalyzer"
	esanalyzerWFTypeName               = "cadence-sys-es-analyzer-workflow"
	emitWorkflowVersionMetricsActivity = "cadence-sys-es-analyzer-emit-workflow-version-metrics"
)

type (
	DomainWorkflowVersionCount struct {
		WorkflowTypes []WorkflowTypeCount `json:"buckets"`
	}
	WorkflowTypeCount struct {
		EsAggregateCount
		WorkflowVersions WorkflowVersionCount `json:"versions"`
	}
	WorkflowVersionCount struct {
		WorkflowVersions []EsAggregateCount `json:"buckets"`
	}
)

var (
	retryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: 10 * time.Minute,
	}

	getWorkflowMetricsOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           esAnalyzerWFID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 10 * time.Minute,
		CronSchedule:                 "*/10 * * * *",
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

// workflowFunc queries ElasticSearch for information and do something with it
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
    "aggs" : {
        "wftypes" : {
            "terms" : { "field" : "WorkflowType"},
            "aggs": {
                "versions": {
                    "terms" : { "field" : "Attr.CadenceChangeVersion"}
                }
            }
        }
    },
    "query": {
        "bool": {
            "must_not": {
                "exists": {
                    "field": "CloseTime"
                }
            },
            "must": [
                {
                    "match" : {
                        "DomainID" : "%s"
                    }
                }
            ]
        }
    },
    "size": 0
}
    `, domain.GetInfo().ID), nil
}

// emitWorkflowVersionMetrics is an activity that emits the running WF versions of a domain
func (w *Workflow) emitWorkflowVersionMetrics(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	var workflowMetricDomainNames []string
	workflowMetricDomains := w.analyzer.config.ESAnalyzerWorkflowVersionDomains()
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
			agg, foundAggregation := response.Aggregations[workflowTypesAggKey]

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
			for _, workflowType := range domainWorkflowVersionCount.WorkflowTypes {
				for _, workflowVersion := range workflowType.WorkflowVersions.WorkflowVersions {
					w.analyzer.tallyScope.Tagged(
						map[string]string{domainTag: domainName, workflowVersionTag: workflowVersion.AggregateKey, workflowTypeTag: workflowType.AggregateKey},
					).Gauge(workflowVersionCountMetrics).Update(float64(workflowVersion.AggregateCount))
				}
			}
		}
	}
	return nil
}

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

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"

	"go.uber.org/cadence"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
)

const (
	workflowTypeCountMetrics = "workflow_type_count"

	// workflow constants
	domainWFTypeCountWorkflowID                = "cadence-sys-tl-esanalyzer-domain-wf-type-count"
	domainWFTypeCountWorkflowTypeName          = "cadence-sys-es-analyzer-domain-wf-type-count-workflow"
	emitDomainWorkflowTypeCountMetricsActivity = "cadence-sys-es-analyzer-emit-domain-workflow-type-count-metrics"
)

type (
	DomainWorkflowTypeCount struct {
		WorkflowTypes []EsAggregateCount `json:"buckets"`
	}
)

var (
	workflowActivityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    10 * time.Second,
			BackoffCoefficient: 1.7,
			MaximumInterval:    5 * time.Minute,
			ExpirationInterval: 10 * time.Minute,
		},
	}

	domainWfTypeCountStartOptions = cclient.StartWorkflowOptions{
		ID:                           domainWFTypeCountWorkflowID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 5 * time.Minute,
		CronSchedule:                 "*/5 * * * *",
	}
)

func initDomainWorkflowTypeCountWorkflow(a *Analyzer) {
	w := Workflow{analyzer: a}
	workflow.RegisterWithOptions(w.emitWorkflowTypeCount, workflow.RegisterOptions{Name: domainWFTypeCountWorkflowTypeName})
	activity.RegisterWithOptions(
		w.emitWorkflowTypeCountMetrics,
		activity.RegisterOptions{Name: emitDomainWorkflowTypeCountMetricsActivity},
	)
}

// emitWorkflowTypeCount queries ElasticSearch for workflow count per type and emit metrics
func (w *Workflow) emitWorkflowTypeCount(ctx workflow.Context) error {
	if w.analyzer.config.ESAnalyzerPause() {
		logger := workflow.GetLogger(ctx)
		logger.Info("Skipping ESAnalyzer execution cycle since it was paused")
		return nil
	}
	var err error
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, workflowActivityOptions),
		emitDomainWorkflowTypeCountMetricsActivity,
	).Get(ctx, &err)
	if err != nil {
		return err
	}
	return nil
}

// get open workflows count per workflow type for specified domain
func (w *Workflow) getDomainWorkflowTypeCountQuery(domainName string) (string, error) {
	domain, err := w.analyzer.domainCache.GetDomain(domainName)
	if err != nil {
		return "", err
	}
	// exclude uninitialized workflow executions by checking whether record has start time field
	return fmt.Sprintf(`
{
    "aggs" : {
        "wftypes" : {
            "terms" : { "field" : "WorkflowType"}
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
                },
				{
					"exists": {
						"field": "StartTime" 
					}
				}
            ]
        }
    },
    "size": 0
}
    `, domain.GetInfo().ID), nil
}

// emitWorkflowTypeCountMetrics is an activity that emits the running workflow type counts of a domain
func (w *Workflow) emitWorkflowTypeCountMetrics(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	var workflowMetricDomainNames []string
	workflowMetricDomains := w.analyzer.config.ESAnalyzerWorkflowTypeDomains()
	if len(workflowMetricDomains) > 0 {
		err := json.Unmarshal([]byte(workflowMetricDomains), &workflowMetricDomainNames)
		if err != nil {
			return err
		}
		for _, domainName := range workflowMetricDomainNames {
			wfTypeCountEsQuery, err := w.getDomainWorkflowTypeCountQuery(domainName)
			if err != nil {
				logger.Error("Failed to get ElasticSearch query to find domain workflow type Info",
					zap.Error(err),
					zap.String("DomainName", domainName),
				)
				return err
			}
			response, err := w.analyzer.esClient.SearchRaw(ctx, w.analyzer.visibilityIndexName, wfTypeCountEsQuery)
			if err != nil {
				logger.Error("Failed to query ElasticSearch to find workflow type count Info",
					zap.Error(err),
					zap.String("VisibilityQuery", wfTypeCountEsQuery),
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
					zap.String("VisibilityQuery", wfTypeCountEsQuery),
				)
				return err
			}
			var domainWorkflowTypeCount DomainWorkflowTypeCount
			err = json.Unmarshal(agg, &domainWorkflowTypeCount)
			if err != nil {
				logger.Error("ElasticSearch error parsing aggregation.",
					zap.Error(err),
					zap.String("Aggregation", string(agg)),
					zap.String("DomainName", domainName),
					zap.String("VisibilityQuery", wfTypeCountEsQuery),
				)
				return err
			}
			for _, workflowType := range domainWorkflowTypeCount.WorkflowTypes {
				w.analyzer.tallyScope.Tagged(
					map[string]string{domainTag: domainName, workflowTypeTag: workflowType.AggregateKey},
				).Gauge(workflowTypeCountMetrics).Update(float64(workflowType.AggregateCount))
			}
		}
	}
	return nil
}

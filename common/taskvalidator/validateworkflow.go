// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package taskvalidator provides a Work in Progress service for workflow validations.
package taskvalidator

import (
	"context"

	"go.uber.org/zap"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/types"
)

type Checker interface {
	WorkflowCheckforValidation(ctx context.Context, workflowID string, domainID string, domainName string, runID string) error
}

type checkerImpl struct {
	logger        log.Logger
	metricsClient metrics.Client
	dc            cache.DomainCache
	pr            persistence.Retryer
	staleCheck    *invariant.StaleWorkflowCheck
}

func NewWfChecker(logger log.Logger, metrics metrics.Client, domainCache cache.DomainCache, pr persistence.Retryer) Checker {
	zapLogger, _ := zap.NewProduction()
	staleCheck := invariant.NewStaleWorkflow(pr, domainCache, zapLogger).(*invariant.StaleWorkflowCheck) // Type assert to *StaleWorkflowCheck
	return &checkerImpl{
		logger:        logger,
		metricsClient: metrics,
		dc:            domainCache,
		pr:            pr,
		staleCheck:    staleCheck,
	}
}

func (w *checkerImpl) WorkflowCheckforValidation(ctx context.Context, workflowID string, domainID string, domainName string, runID string) error {
	w.logger.Info("WorkflowCheckforValidation",
		tag.WorkflowID(workflowID),
		tag.WorkflowRunID(runID),
		tag.WorkflowDomainID(domainID),
		tag.WorkflowDomainName(domainName))

	workflowResp, err := w.pr.GetWorkflowExecution(ctx, &persistence.GetWorkflowExecutionRequest{
		DomainID: domainID,
		Execution: types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	})
	if err != nil {
		w.logger.Error("Error getting workflow execution", tag.Error(err))
		return err
	}

	if w.isDomainDeprecated(domainID) {
		invariant.DeleteExecution(ctx, workflowResp.State.ExecutionInfo, w.pr, w.dc)
		return nil
	}

	if w.staleWorkflowCheck(workflowResp) {
		invariant.DeleteExecution(ctx, workflowResp.State.ExecutionInfo, w.pr, w.dc)
		return nil
	}

	w.metricsClient.Scope(metrics.TaskValidatorScope, metrics.DomainTag(domainName)).IncCounter(metrics.ValidatedWorkflowCount)
	return nil
}

func (w *checkerImpl) isDomainDeprecated(domainID string) bool {
	domain, err := w.dc.GetDomainByID(domainID)
	if err != nil {
		w.logger.Error("Error getting domain by ID", tag.Error(err))
		return false
	}
	return domain.IsDeprecatedOrDeleted()
}

func (w *checkerImpl) staleWorkflowCheck(workflowResp *persistence.GetWorkflowExecutionResponse) bool {
	pastExpiration, _ := w.staleCheck.CheckAge(workflowResp)
	return pastExpiration
}

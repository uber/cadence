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
	"fmt"

	"github.com/uber/cadence/common/log"
)

// Checker is an interface for initiating the validation process.
type Checker interface {
	WorkflowCheckforValidation(workflowID string, domainID string, runID string) error
}

// checkerImpl is the implementation of the Checker interface.
type checkerImpl struct {
	logger log.Logger
}

// NewWfChecker creates a new instance of Checker.
func NewWfChecker(logger log.Logger) Checker {
	return &checkerImpl{logger: logger}
}

// WorkflowCheckforValidation is a dummy implementation of workflow validation.
func (w *checkerImpl) WorkflowCheckforValidation(workflowID string, domainID string, runID string) error {
	// Emitting just the log to ensure that the workflow is called for now.
	// TODO: add some validations to check the wf for corruptions.
	w.logger.Info(fmt.Sprintf("WorkflowCheckforValidation. DomainID: %v, WorkflowID: %v, RunID: %v",
		domainID,
		workflowID,
		runID))
	return nil
}

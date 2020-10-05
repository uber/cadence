// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package thrift

import (
	"time"

	thrift "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

// FromWorkflowExecution converts internal WorkflowExectution type to thrift
func FromWorkflowExecution(we types.WorkflowExecution) *thrift.WorkflowExecution {
	return &thrift.WorkflowExecution{
		WorkflowId: &we.WorkflowId,
		RunId:      &we.RunId,
	}
}

// ToWorkflowExecution converts thrift WorkflowExectution type to internal
func ToWorkflowExecution(we *thrift.WorkflowExecution) types.WorkflowExecution {
	if we == nil {
		return types.WorkflowExecution{}
	}
	return types.WorkflowExecution{
		WorkflowId: we.GetWorkflowId(),
		RunId:      we.GetRunId(),
	}
}

// FromWorkflowType converts internal WorkflowType type to thrift
func FromWorkflowType(wt types.WorkflowType) *thrift.WorkflowType {
	return &thrift.WorkflowType{
		Name: &wt.Name,
	}
}

// ToWorkflowType converts thrift WorkflowType type to internal
func ToWorkflowType(wt *thrift.WorkflowType) types.WorkflowType {
	if wt == nil {
		return types.WorkflowType{}
	}
	return types.WorkflowType{
		Name: wt.GetName(),
	}
}

// FromActivityType converts internal ActivityType type to thrift
func FromActivityType(at types.ActivityType) *thrift.ActivityType {
	return &thrift.ActivityType{
		Name: &at.Name,
	}
}

// ToActivityType converts thrift ActivityType type to internal
func ToActivityType(at *thrift.ActivityType) types.ActivityType {
	if at == nil {
		return types.ActivityType{}
	}
	return types.ActivityType{
		Name: at.GetName(),
	}
}

// FromHeader converts internal Header type to thrift
func FromHeader(h *types.Header) *thrift.Header {
	if h == nil {
		return nil
	}
	return &thrift.Header{
		Fields: h.Fields,
	}
}

// ToHeader converts thrift Header type to internal
func ToHeader(h *thrift.Header) *types.Header {
	if h == nil {
		return nil
	}
	return &types.Header{
		Fields: h.Fields,
	}
}

// FromMemo converts internal Memo type to thrift
func FromMemo(m *types.Memo) *thrift.Memo {
	if m == nil {
		return nil
	}
	return &thrift.Memo{
		Fields: m.Fields,
	}
}

// ToMemo converts thrift Memo type to internal
func ToMemo(m *thrift.Memo) *types.Memo {
	if m == nil {
		return nil
	}
	return &types.Memo{
		Fields: m.Fields,
	}
}

// FromSearchAttributes converts internal SearchAttributes type to thrift
func FromSearchAttributes(sa *types.SearchAttributes) *thrift.SearchAttributes {
	if sa == nil {
		return nil
	}
	return &thrift.SearchAttributes{
		IndexedFields: sa.IndexedFields,
	}
}

// ToSearchAttributes converts thrift SearchAttributes type to internal
func ToSearchAttributes(sa *thrift.SearchAttributes) *types.SearchAttributes {
	if sa == nil {
		return nil
	}
	return &types.SearchAttributes{
		IndexedFields: sa.IndexedFields,
	}
}

// FromRetryPolicy converts internal RetryPolicy type to thrift
func FromRetryPolicy(rp *types.RetryPolicy) *thrift.RetryPolicy {
	if rp == nil {
		return nil
	}
	return &thrift.RetryPolicy{
		InitialIntervalInSeconds:    durationToSeconds(rp.InitialInterval),
		BackoffCoefficient:          &rp.BackoffCoefficient,
		MaximumIntervalInSeconds:    durationToSeconds(rp.MaximumInterval),
		MaximumAttempts:             &rp.MaximumAttempts,
		NonRetriableErrorReasons:    rp.NonRetriableErrorReasons,
		ExpirationIntervalInSeconds: durationToSeconds(rp.ExpirationInterval),
	}
}

// ToRetryPolicy converts thrift RetryPolicy type to internal
func ToRetryPolicy(rp *thrift.RetryPolicy) *types.RetryPolicy {
	if rp == nil {
		return nil
	}
	return &types.RetryPolicy{
		InitialInterval:          secondsToDuration(rp.GetInitialIntervalInSeconds()),
		BackoffCoefficient:       rp.GetBackoffCoefficient(),
		MaximumInterval:          secondsToDuration(rp.GetMaximumIntervalInSeconds()),
		MaximumAttempts:          rp.GetMaximumAttempts(),
		NonRetriableErrorReasons: rp.NonRetriableErrorReasons,
		ExpirationInterval:       secondsToDuration(rp.GetExpirationIntervalInSeconds()),
	}
}

func durationToSeconds(d time.Duration) *int32 {
	seconds := int32(d.Seconds())
	return &seconds
}

func secondsToDuration(s int32) time.Duration {
	return time.Duration(s) * time.Second
}

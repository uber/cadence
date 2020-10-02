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
		WorkflowId: stringPtr(we.WorkflowId),
		RunId:      stringPtr(we.RunId),
	}
}

// ToWorkflowExecution converts thrift WorkflowExectution type to internal
func ToWorkflowExecution(we *thrift.WorkflowExecution) types.WorkflowExecution {
	if we == nil {
		return types.WorkflowExecution{}
	}
	return types.WorkflowExecution{
		WorkflowId: stringVal(we.WorkflowId),
		RunId:      stringVal(we.RunId),
	}
}

// FromWorkflowType converts internal WorkflowType type to thrift
func FromWorkflowType(wt types.WorkflowType) *thrift.WorkflowType {
	return &thrift.WorkflowType{
		Name: stringPtr(wt.Name),
	}
}

// ToWorkflowType converts thrift WorkflowType type to internal
func ToWorkflowType(wt *thrift.WorkflowType) types.WorkflowType {
	if wt == nil {
		return types.WorkflowType{}
	}
	return types.WorkflowType{
		Name: stringVal(wt.Name),
	}
}

// FromActivityType converts internal ActivityType type to thrift
func FromActivityType(at types.ActivityType) *thrift.ActivityType {
	return &thrift.ActivityType{
		Name: stringPtr(at.Name),
	}
}

// ToActivityType converts thrift ActivityType type to internal
func ToActivityType(at *thrift.ActivityType) types.ActivityType {
	if at == nil {
		return types.ActivityType{}
	}
	return types.ActivityType{
		Name: stringVal(at.Name),
	}
}

// FromHeader converts internal Header type to thrift
func FromHeader(h types.Header) *thrift.Header {
	return &thrift.Header{
		Fields: h.Fields,
	}
}

// ToHeader converts thrift Header type to internal
func ToHeader(h *thrift.Header) types.Header {
	if h == nil {
		return types.Header{}
	}
	return types.Header{
		Fields: h.Fields,
	}
}

// FromMemo converts internal Memo type to thrift
func FromMemo(h types.Memo) *thrift.Memo {
	return &thrift.Memo{
		Fields: h.Fields,
	}
}

// ToMemo converts thrift Memo type to internal
func ToMemo(h *thrift.Memo) types.Memo {
	if h == nil {
		return types.Memo{}
	}
	return types.Memo{
		Fields: h.Fields,
	}
}

// FromSearchAttributes converts internal SearchAttributes type to thrift
func FromSearchAttributes(sa types.SearchAttributes) *thrift.SearchAttributes {
	return &thrift.SearchAttributes{
		IndexedFields: sa.IndexedFields,
	}
}

// ToSearchAttributes converts thrift SearchAttributes type to internal
func ToSearchAttributes(sa *thrift.SearchAttributes) types.SearchAttributes {
	if sa == nil {
		return types.SearchAttributes{}
	}
	return types.SearchAttributes{
		IndexedFields: sa.IndexedFields,
	}
}

// FromRetryPolicy converts internal RetryPolicy type to thrift
func FromRetryPolicy(rp *types.RetryPolicy) *thrift.RetryPolicy {
	if rp == nil {
		return nil
	}
	return &thrift.RetryPolicy{
		InitialIntervalInSeconds:    int32Ptr(int32(rp.InitialInterval.Seconds())),
		BackoffCoefficient:          float64Ptr(rp.BackoffCoefficient),
		MaximumIntervalInSeconds:    int32Ptr(int32(rp.MaximumInterval.Seconds())),
		MaximumAttempts:             int32Ptr(rp.MaximumAttempts),
		NonRetriableErrorReasons:    rp.NonRetriableErrorReasons,
		ExpirationIntervalInSeconds: int32Ptr(int32(rp.ExpirationInterval.Seconds())),
	}
}

// ToRetryPolicy converts thrift RetryPolicy type to internal
func ToRetryPolicy(rp *thrift.RetryPolicy) *types.RetryPolicy {
	if rp == nil {
		return nil
	}
	return &types.RetryPolicy{
		InitialInterval:          time.Second * time.Duration(int32Val(rp.InitialIntervalInSeconds)),
		BackoffCoefficient:       float64Val(rp.BackoffCoefficient),
		MaximumInterval:          time.Second * time.Duration(int32Val(rp.MaximumIntervalInSeconds)),
		MaximumAttempts:          int32Val(rp.MaximumAttempts),
		NonRetriableErrorReasons: rp.NonRetriableErrorReasons,
		ExpirationInterval:       time.Second * time.Duration(int32Val(rp.ExpirationIntervalInSeconds)),
	}
}

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	thrift "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

func Test_FromWorkflowExecution(t *testing.T) {
	assert.Equal(t,
		&thrift.WorkflowExecution{
			WorkflowId: stringPtr(""),
			RunId:      stringPtr(""),
		},
		FromWorkflowExecution(types.WorkflowExecution{}),
	)

	assert.Equal(t,
		&thrift.WorkflowExecution{
			WorkflowId: stringPtr("WorkflowId"),
			RunId:      stringPtr("RunId"),
		},
		FromWorkflowExecution(types.WorkflowExecution{
			WorkflowId: "WorkflowId",
			RunId:      "RunId",
		}),
	)
}

func Test_ToWorkflowExecution(t *testing.T) {
	assert.Equal(t,
		types.WorkflowExecution{},
		ToWorkflowExecution(nil),
	)

	assert.Equal(t,
		types.WorkflowExecution{},
		ToWorkflowExecution(&thrift.WorkflowExecution{}),
	)

	assert.Equal(t,
		types.WorkflowExecution{
			WorkflowId: "WorkflowId",
			RunId:      "RunId",
		},
		ToWorkflowExecution(&thrift.WorkflowExecution{
			WorkflowId: stringPtr("WorkflowId"),
			RunId:      stringPtr("RunId"),
		}),
	)
}

func Test_FromWorkflowType(t *testing.T) {
	assert.Equal(t, &thrift.WorkflowType{Name: stringPtr("")}, FromWorkflowType(types.WorkflowType{}))
	assert.Equal(t, &thrift.WorkflowType{Name: stringPtr("Name")}, FromWorkflowType(types.WorkflowType{Name: "Name"}))
}

func Test_ToWorkflowType(t *testing.T) {
	assert.Equal(t, types.WorkflowType{}, ToWorkflowType(nil))
	assert.Equal(t, types.WorkflowType{}, ToWorkflowType(&thrift.WorkflowType{}))
	assert.Equal(t, types.WorkflowType{Name: "Name"}, ToWorkflowType(&thrift.WorkflowType{Name: stringPtr("Name")}))
}

func Test_FromActivityType(t *testing.T) {
	assert.Equal(t, &thrift.ActivityType{Name: stringPtr("")}, FromActivityType(types.ActivityType{}))
	assert.Equal(t, &thrift.ActivityType{Name: stringPtr("Name")}, FromActivityType(types.ActivityType{Name: "Name"}))
}

func Test_ToActivityType(t *testing.T) {
	assert.Equal(t, types.ActivityType{}, ToActivityType(nil))
	assert.Equal(t, types.ActivityType{}, ToActivityType(&thrift.ActivityType{}))
	assert.Equal(t, types.ActivityType{Name: "Name"}, ToActivityType(&thrift.ActivityType{Name: stringPtr("Name")}))
}

func Test_FromHeader(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, &thrift.Header{}, FromHeader(types.Header{}))
	assert.Equal(t, &thrift.Header{Fields: fields}, FromHeader(types.Header{Fields: fields}))
}

func Test_ToHeader(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, types.Header{}, ToHeader(nil))
	assert.Equal(t, types.Header{}, ToHeader(&thrift.Header{}))
	assert.Equal(t, types.Header{Fields: fields}, ToHeader(&thrift.Header{Fields: fields}))
}

func Test_FromMemo(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, &thrift.Memo{}, FromMemo(types.Memo{}))
	assert.Equal(t, &thrift.Memo{Fields: fields}, FromMemo(types.Memo{Fields: fields}))
}

func Test_ToMemo(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, types.Memo{}, ToMemo(nil))
	assert.Equal(t, types.Memo{}, ToMemo(&thrift.Memo{}))
	assert.Equal(t, types.Memo{Fields: fields}, ToMemo(&thrift.Memo{Fields: fields}))
}

func Test_FromSearchAttributes(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, &thrift.SearchAttributes{}, FromSearchAttributes(types.SearchAttributes{}))
	assert.Equal(t, &thrift.SearchAttributes{IndexedFields: fields}, FromSearchAttributes(types.SearchAttributes{IndexedFields: fields}))
}

func Test_ToSearchAttributes(t *testing.T) {
	fields := map[string][]byte{"A": {1, 2, 3}}
	assert.Equal(t, types.SearchAttributes{}, ToSearchAttributes(nil))
	assert.Equal(t, types.SearchAttributes{}, ToSearchAttributes(&thrift.SearchAttributes{}))
	assert.Equal(t, types.SearchAttributes{IndexedFields: fields}, ToSearchAttributes(&thrift.SearchAttributes{IndexedFields: fields}))
}

func Test_FromRetryPolicy(t *testing.T) {
	assert.Equal(t, (*thrift.RetryPolicy)(nil), FromRetryPolicy(nil))

	assert.Equal(t,
		&thrift.RetryPolicy{
			InitialIntervalInSeconds:    int32Ptr(0),
			BackoffCoefficient:          float64Ptr(0.0),
			MaximumIntervalInSeconds:    int32Ptr(0),
			MaximumAttempts:             int32Ptr(0),
			NonRetriableErrorReasons:    nil,
			ExpirationIntervalInSeconds: int32Ptr(0),
		},
		FromRetryPolicy(&types.RetryPolicy{}))

	assert.Equal(t,
		&thrift.RetryPolicy{
			InitialIntervalInSeconds:    int32Ptr(1),
			BackoffCoefficient:          float64Ptr(2.0),
			MaximumIntervalInSeconds:    int32Ptr(3),
			MaximumAttempts:             int32Ptr(4),
			NonRetriableErrorReasons:    []string{"5"},
			ExpirationIntervalInSeconds: int32Ptr(6),
		},
		FromRetryPolicy(&types.RetryPolicy{
			InitialInterval:          1 * time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          3 * time.Second,
			MaximumAttempts:          4,
			NonRetriableErrorReasons: []string{"5"},
			ExpirationInterval:       6 * time.Second,
		}))
}

func Test_ToRetryPolicy(t *testing.T) {
	assert.Equal(t, (*types.RetryPolicy)(nil), ToRetryPolicy(nil))

	assert.Equal(t,
		&types.RetryPolicy{
			InitialInterval:          0 * time.Second,
			BackoffCoefficient:       0.0,
			MaximumInterval:          0 * time.Second,
			MaximumAttempts:          0,
			NonRetriableErrorReasons: nil,
			ExpirationInterval:       0 * time.Second,
		},
		ToRetryPolicy(&thrift.RetryPolicy{}))

	assert.Equal(t,
		&types.RetryPolicy{
			InitialInterval:          1 * time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          3 * time.Second,
			MaximumAttempts:          4,
			NonRetriableErrorReasons: []string{"5"},
			ExpirationInterval:       6 * time.Second,
		},
		ToRetryPolicy(&thrift.RetryPolicy{
			InitialIntervalInSeconds:    int32Ptr(1),
			BackoffCoefficient:          float64Ptr(2.0),
			MaximumIntervalInSeconds:    int32Ptr(3),
			MaximumAttempts:             int32Ptr(4),
			NonRetriableErrorReasons:    []string{"5"},
			ExpirationIntervalInSeconds: int32Ptr(6),
		}))
}

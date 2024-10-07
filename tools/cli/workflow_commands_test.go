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

package cli

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestConstructStartWorkflowRequest(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "domain")
	set.String(FlagTaskList, "test-task-list", "tasklist")
	set.String(FlagWorkflowType, "test-workflow-type", "workflow-type")
	set.Int("execution_timeout", 100, "execution_timeout")
	set.Int("decision_timeout", 50, "decision_timeout")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Int("workflow_id_reuse_policy", 1, "workflow_id_reuse_policy")
	set.String("input", "{}", "input")
	set.String("cron_schedule", "* * * * *", "cron_schedule")
	set.Int("retry_attempts", 5, "retry_attempts")
	set.Int("retry_expiration", 600, "retry_expiration")
	set.Int("retry_interval", 10, "retry_interval")
	set.Float64("retry_backoff", 2.0, "retry_backoff")
	set.Int("retry_max_interval", 100, "retry_max_interval")
	set.Int(DelayStartSeconds, 5, DelayStartSeconds)
	set.Int(JitterStartSeconds, 2, JitterStartSeconds)
	set.String("first_run_at_time", "2024-07-24T12:00:00Z", "first-run-at-time")

	c := cli.NewContext(nil, set, nil)
	// inject context with span
	tracer := mocktracer.New()
	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test-span")
	span.SetBaggageItem("tracer-test-key", "tracer-test-value")
	defer span.Finish()
	c.Context = ctx

	assert.NoError(t, c.Set(FlagDomain, "test-domain"))
	assert.NoError(t, c.Set(FlagTaskList, "test-task-list"))
	assert.NoError(t, c.Set(FlagWorkflowType, "test-workflow-type"))
	assert.NoError(t, c.Set("execution_timeout", "100"))
	assert.NoError(t, c.Set("decision_timeout", "50"))
	assert.NoError(t, c.Set("workflow_id", "test-workflow-id"))
	assert.NoError(t, c.Set("workflow_id_reuse_policy", "1"))
	assert.NoError(t, c.Set("input", "{}"))
	assert.NoError(t, c.Set("cron_schedule", "* * * * *"))
	assert.NoError(t, c.Set("retry_attempts", "5"))
	assert.NoError(t, c.Set("retry_expiration", "600"))
	assert.NoError(t, c.Set("retry_interval", "10"))
	assert.NoError(t, c.Set("retry_backoff", "2.0"))
	assert.NoError(t, c.Set("retry_max_interval", "100"))
	assert.NoError(t, c.Set(DelayStartSeconds, "5"))
	assert.NoError(t, c.Set(JitterStartSeconds, "2"))
	assert.NoError(t, c.Set("first_run_at_time", "2024-07-24T12:00:00Z"))

	request, err := constructStartWorkflowRequest(c)
	assert.NoError(t, err)
	assert.NotNil(t, request)
	assert.Equal(t, "test-domain", request.Domain)
	assert.Equal(t, "test-task-list", request.TaskList.Name)
	assert.Equal(t, "test-workflow-type", request.WorkflowType.Name)
	assert.Equal(t, int32(100), *request.ExecutionStartToCloseTimeoutSeconds)
	assert.Equal(t, int32(50), *request.TaskStartToCloseTimeoutSeconds)
	assert.Equal(t, "test-workflow-id", request.WorkflowID)
	assert.NotNil(t, request.WorkflowIDReusePolicy)
	assert.Equal(t, int32(5), *request.DelayStartSeconds)
	assert.Equal(t, int32(2), *request.JitterStartSeconds)
	assert.Equal(t, map[string][]byte{"tracer-test-key": []byte("tracer-test-value")}, request.Header.Fields)

	firstRunAt, err := time.Parse(time.RFC3339, "2024-07-24T12:00:00Z")
	assert.NoError(t, err)
	assert.Equal(t, firstRunAt.UnixNano(), *request.FirstRunAtTimeStamp)
}

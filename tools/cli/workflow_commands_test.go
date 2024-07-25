package cli

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"testing"
	"time"
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

	c.Set(FlagDomain, "test-domain")
	c.Set(FlagTaskList, "test-task-list")
	c.Set(FlagWorkflowType, "test-workflow-type")
	c.Set("execution_timeout", "100")
	c.Set("decision_timeout", "50")
	c.Set("workflow_id", "test-workflow-id")
	c.Set("workflow_id_reuse_policy", "1")
	c.Set("input", "{}")
	c.Set("cron_schedule", "* * * * *")
	c.Set("retry_attempts", "5")
	c.Set("retry_expiration", "600")
	c.Set("retry_interval", "10")
	c.Set("retry_backoff", "2.0")
	c.Set("retry_max_interval", "100")
	c.Set(DelayStartSeconds, "5")
	c.Set(JitterStartSeconds, "2")
	c.Set("first_run_at_time", "2024-07-24T12:00:00Z")

	request := constructStartWorkflowRequest(c)

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

	firstRunAt, err := time.Parse(time.RFC3339, "2024-07-24T12:00:00Z")
	assert.NoError(t, err)
	assert.Equal(t, firstRunAt.UnixNano(), *request.FirstRunAtTimeStamp)
}

---
layout: default
title: Queries
permalink: /docs/go-client/queries
---

# Queries

If a :workflow_execution: has been stuck at a state for longer than an expected period of time, you
might want to :query: the current call stack. You can use the Cadence :CLI: to perform this :query:. For
example:

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt __stack_trace`

This command uses `__stack_trace`, which is a built-in :query: type supported by the Cadence client
library. You can add custom :query: types to handle :query:queries: such as :query:querying: the current state of a
:workflow:, or :query:querying: how many :activity:activities: the :workflow: has completed. To do this, you need to set
up a :query: handler using `workflow.SetQueryHandler`.

The handler must be a function that returns two values:
1. A serializable result
2. An error

The handler function can receive any number of input parameters, but all input parameters must be
serializable. The following sample code sets up a :query: handler that handles the :query: type of
`current_state`:
```go
func MyWorkflow(ctx workflow.Context, input string) error {
    currentState := "started" // This could be any serializable struct.
    err := workflow.SetQueryHandler(ctx, "current_state", func() (string, error) {
        return currentState, nil
    })
    if err != nil {
        currentState = "failed to register query handler"
        return err
    }
    // Your normal workflow code begins here, and you update the currentState as the code makes progress.
    currentState = "waiting timer"
    err = NewTimer(ctx, time.Hour).Get(ctx, nil)
    if err != nil {
        currentState = "timer failed"
        return err
    }

    currentState = "waiting activity"
    ctx = WithActivityOptions(ctx, myActivityOptions)
    err = ExecuteActivity(ctx, MyActivity, "my_input").Get(ctx, nil)
    if err != nil {
        currentState = "activity failed"
        return err
    }
    currentState = "done"
    return nil
}
```
You can now :query: `current_state` by using the :CLI::

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state`

You can also issue a :query: from code using the `QueryWorkflow()` API on a Cadence client object.

## Consistent Query

:query:Query: has two consistency levels, eventual and strong. Consider if you were to :signal: a :workflow: and then
immediately :query: the :workflow::

`cadence-cli --domain samples-domain workflow signal -w my_workflow_id -r my_run_id -n signal_name -if ./input.json`

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state`

In this example if :signal: were to change :workflow: state, :query: may or may not see that state update reflected
in the :query: result. This is what it means for :query: to be eventually consistent.

:query:Query: has another consistency level called strong consistency. A strongly consistent :query: is guaranteed
to be based on :workflow: state which includes all :event:events: that came before the :query: was issued. An :event:
is considered to have come before a :query: if the call creating the external :event: returned success before
the :query: was issued. External :event:events: which are created while the :query: is outstanding may or may not
be reflected in the :workflow: state the :query: result is based on.

In order to run consistent :query: through the :CLI: do the following:

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state --qcl strong`

In order to run a :query: using the go client do the following:

```go
resp, err := cadenceClient.QueryWorkflowWithOptions(ctx, &client.QueryWorkflowWithOptionsRequest{
    WorkflowID:            workflowID,
    RunID:                 runID,
    QueryType:             queryType,
    QueryConsistencyLevel: shared.QueryConsistencyLevelStrong.Ptr(),
})
```

When using strongly consistent :query: you should expect higher latency than eventually consistent :query:.

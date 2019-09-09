# Queries

If a workflow execution has been stuck at a state for longer than an expected period of time, you
might want to query the current call stack. You can use the Cadence CLI to perform this query. For
example:

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt __stack_trace`

This command uses `__stack_trace`, which is a built-in query type supported by the Cadence client
library. You can add custom query types to handle queries such as querying the current state of a
workflow, or querying how many activities the workflow has completed. To do this, you need to set
up a query handler using `workflow.SetQueryHandler`.

The handler must be a function that returns two values:
1. A serializable result
2. An error

The handler function can receive any number of input parameters, but all input parameters must be
serializable. The following sample code sets up a query handler that handles the query type of
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
You can now query `current_state` by using the CLI:

`cadence-cli --domain samples-domain workflow query -w my_workflow_id -r my_run_id -qt current_state`

You can also issue a query from code using the `QueryWorkflow()` API on a Cadence client object.

# Child workflows

`workflow.ExecuteChildWorkflow` enables the scheduling of other workflows from within a workflow's
implementation. The parent workflow has the ability to monitor and impact the lifecycle of the child
workflow, similar to the way it does for an activity that it invoked.

```go
cwo := workflow.ChildWorkflowOptions{
        // Do not specify WorkflowID if you want cadence to generate a unique ID for the child execution.
        WorkflowID:                   "BID-SIMPLE-CHILD-WORKFLOW",
        ExecutionStartToCloseTimeout: time.Minute * 30,
}
ctx = workflow.WithChildWorkflowOptions(ctx, cwo)

var result string
future := workflow.ExecuteChildWorkflow(ctx, SimpleChildWorkflow, value)
if err := future.Get(ctx, &result); err != nil {
        workflow.GetLogger(ctx).Error("SimpleChildWorkflow failed.", zap.Error(err))
        return err
}
```
Let's take a look at each component of this call.

Before calling `workflow.ExecuteChildworkflow()`, you must configure `ChildWorkflowOptions` for the
invocation. These options customize various execution timeouts, and are passed in by creating a child
context from the initial context and overwriting the desired values. The child context is then passed
into the `workflow.ExecuteChildWorkflow()` call. If multiple activities are sharing the same option
values, then the same context instance can be used when calling `workflow.ExecuteChildworkflow()`.

The first parameter in the call is the required `cadence.Context` object. This type is a copy of
`context.Context` with the `Done()` method returning `cadence.Channel` instead of the native Go `chan`.

The second parameter is the function that we registered as a workflow function. This parameter can
also be a string representing the fully qualified name of the workflow function. The benefit of this
is that when you pass in the actual function object, the framework can validate workflow parameters.

The remaining parameters are passed to the workflow as part of the call. In our example, we have a
single parameter: `value`. This list of parameters must match the list of parameters declared by
the workflow function.

The method call returns immediately and returns a `cadence.Future`. This allows you to execute more
code without having to wait for the scheduled workflow to complete.

When you are ready to process the results of the workflow, call the `Get()` method on the future
object returned. The parameters to this method is the `ctx` object we passed to the
`workflow.ExecuteChildWorkflow()` call and an output parameter that will receive the output of the
workflow. The type of the output parameter must match the type of the return value declared by the
workflow function. The `Get()` method will block until the workflow completes and results are
available.

The `workflow.ExecuteChildWorkflow()` function is similar to `workflow.ExecuteActivity()`. All of the
patterns described for using `workflow.ExecuteActivity()` apply to the `workflow.ExecuteChildWorkflow()`
function as well.

When a parent workflow is cancelled by the user, the child workflow can be cancelled or abandoned
based on a configurable child policy.

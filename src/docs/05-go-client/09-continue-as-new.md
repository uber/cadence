---
layout: default
title: Continue as new
permalink: /docs/go-client/continue-as-new
---

# Continue as new

:workflow:Workflows: that need to rerun periodically could naively be implemented as a big **for** loop with
a sleep where the entire logic of the :workflow: is inside the body of the **for** loop. The problem
with this approach is that the history for that :workflow: will keep growing to a point where it
reaches the maximum size enforced by the service.

**ContinueAsNew** is the low level construct that enables implementing such :workflow:workflows: without the
risk of failures down the road. The operation atomically completes the current execution and starts
a new execution of the :workflow: with the same **:workflow_ID:**. The new execution will not carry
over any history from the old execution. To trigger this behavior, the :workflow: function should
terminate by returning the special **ContinueAsNewError** error:

```go
func SimpleWorkflow(workflow.Context ctx, value string) error {
    ...
    return workflow.NewContinueAsNewError(ctx, SimpleWorkflow, value)
}
```

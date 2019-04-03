
# Signals

**Signals** provide a mechanism to send data directly to a running workflow. Previously, you had 
two options for passing data to the workflow implementation:

* Via start parameters
* As return values from activities

With start parameters, we could only pass in values before workflow execution begins.

Return values from activities allowed us to pass information to a running workflow, but this 
approach comes with its own complications. One major drawback is reliance on polling. This means 
that the data needs to be stored in a third-party location until it's ready to be picked up by 
the activity. Further, the lifecycle of this activity requires management, and the activity 
requires manual restart if it fails before acquiring the data.

**Signals**, on the other hand, provide a fully async and durable mechanism for providing data to
a running workflow. When a signal is received for a running workflow, Cadence persists the event 
and the payload in the workflow history. The workflow can then process the signal at any time 
afterwards without the risk of losing the information. The workflow also has the option to stop 
execution by blocking on a **signal channel**.

```go
var signalVal string
signalChan := workflow.GetSignalChannel(ctx, signalName)

s := workflow.NewSelector(ctx)
s.AddReceive(signalChan, func(c workflow.Channel, more bool) {
    c.Receive(ctx, &signalVal)
    workflow.GetLogger(ctx).Info("Received signal!", zap.String("signal", signalName), zap.String("value", signalVal))
})
s.Select(ctx)

if len(signalVal) > 0 && signalVal != "SOME_VALUE" {
    return errors.New("signalVal")
}
```

In the example above, the workflow code uses **workflow.GetSignalChannel** to open a 
**workflow.Channel** for the named signal. We then use a **workflow.Selector** to wait on this 
channel and process the payload received with the signal.

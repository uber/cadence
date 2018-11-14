---
layout: doc
title: Key Features
weight: 1
---

# Key features

### Dynamic workflow execution graphs
Determine the workflow execution graphs at runtime based on the data you are processing. Cadence
does not pre-compute the execution graphs at compile time or at workflow start time. Therefore, you
have the ability to write workflows that can dynamically adjust to the amount of data they are processing.
If you need to trigger 10 instances of an activity to efficiently process all the data in one run, but
only 3 for a subsequent run, you can do that.

### Child Workflows
Orchestrate the execution of a workflow from within another workflow. Cadence will return the results
of the child workflow execution to the parent workflow upon completion of the child workflow. No polling
is required in the parent workflow to monitor the status of the child workflow, making the process
efficient and fault tolerant.

### Durable Timers
Implement delayed execution of tasks in your workflows that are robust to worker failures. Cadence
provides two easy-to-use APIs for implementing time-based events in your workflows: `cadence.Sleep`
and `cadence.Timer`. Cadence ensures that the timer settings are persisted and the events are generated
even if the workers executing the workflow crash.

### Signals
Modify/influence the execution path of a running workflow by pushing additional data directly to the
workflow using a signal. Via the `Signal` facility, Cadence provides a mechanism to consume external
events directly in workflow code.

### Task routing
Efficiently process large amounts of data using a Cadence workflow by caching the data locally on a
worker and executing all activities meant to process that data on that same worker. Cadence enables
you to choose the worker you want to execute a certain activity by scheduling that activity execution
in the worker's specific task list.

### Unique workflow ID enforcement
Use business entity IDs for your workflows and let Cadence ensure that only one workflow is running
for a particular entity at a time. Cadence implements an atomic "uniqueness check" and ensures that
no race conditions are possible that would result in multiple workflow executions for the same workflow
ID. Therefore, you can implement your code to attempt to start a workflow without checking if the ID
is already in use, even in cases where only one active execution per workflow ID is desired.

### Perpetual/ContinueAsNew workflows
Run periodic tasks as a single perpetually running workflow. With the `ContinueAsNew` facility, Cadence
allows you to leverage the "unique workflow ID enforcement" feature for periodic workflows. Cadence
completes the current execution and starts the new execution atomically, ensuring that you get to keep
your workflow ID. By starting a new execution, Cadence also ensures that workflow execution history
does not grow indefinitely for perpetual workflows.

### At-most-once activity execution
Execute non-idempotent activities as part of your workflows. Cadence will **not** automatically retry
activities on failure. For every activity execution, Cadence will return a success result, a failure
result, or a timeout to the workflow code and let the workflow code determine how each one of those
result types should be handled.

### Asynch Activity Completion
Incorporate human input or thrid-party service asynchronous callbacks into your workflows. Cadence
allows a workflow to pause execution on an activity and wait for an external actor to resume it with
a callback. During this pause, the activity does not have any actively executing code, such as a polling
loop, and is merely an entry in the Cadence datastore. Therefore, the workflow is unaffected by any
worker failures happening over the duration of the pause.

### Activity Heartbeating
Detect unexpected failures or crashes and track progress in long-running activities early. By configuring
your activity to report progress periodically to the Cadence server, you can detect a crash that occurs
10 minutes into an hour-long activity execution much sooner, instead of waiting for the 60-minute execution
timeout. The recorded progress before the crash gives you sufficient information to determine whether
to restart the activity from the beginning or resume it from the point of failure.

### Timeouts for activities and workflow executions
Protect against stuck and unresponsive activities and workflows with appropriate timeout values.
Cadence requires that timeout values are provided for every activity or workflow invocation. There is
no upper bound on the timeout values, so you can set timeouts that span days, weeks, or even months.

### Visibility
Get a list of all your active and/or completed workflows. Explore the execution history of a particular
workflow execution. Cadence provides a set of visibility APIs that allow you, the workflow owner, to
monitor past and current workflow executions.

### Debuggability
Replay any workflow execution history locally under a debugger. The Cadence client library provides
an API that allows you to capture a stack trace from any failed workflow execution history.


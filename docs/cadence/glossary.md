---
layout: doc
title: Glossary
weight: 3
---

# Cadence glossary

This glossary contains terms that are used with the Cadence product.

### Activity
A business-level task that implements your application logic such as calling
a service or transcoding a media file. An activity usually implements a single
well-defined action; it can be short or long running. An activity can be implemented
as a synchronous method or fully asynchronously involving multiple processes. An
activity is executed at most once. This means that the Cadence service never
requests activity execution more than once. If for any reason an activity is not
completed within the specified timeout, an error is reported to the workflow and
the workflow decides how to handle it. There is no limit on potential activity
duration.

### Client Stub
A client-side proxy used to make remote invocations to an entity that it
represents. For example, to start a workflow, a stub object that represents
this workflow is created through a special API. Then this stub is used to start,
query, or signal the corresponding workflow.

### Decision
Any action taken by the workflow function is called a decision. For example:
scheduling a task, canceling a task, or starting a timer.

### Domain
A namespace-like concept. All entities stored in Cadence are stored in a
specific domain. For example, when a workflow is started, it is started in a
specific domain. Cadence guarantees a unique workflow ID within a domain, and
supports running workflow executions to use the same workflow ID if they are in
different domains. Domains are created and updated though a separate CRUD API
or through the CLI.

### Event
An indivisible operation performed by your application. For example,
acitvity_task_started, task_failed, or timer_canceled.

### History
An append log of events for your application. History is durably persisted
by the Cadence service, enabling seamless recovery of your application state
from crashes or failures. It also serves as an audit log for debugging.

### Query
A synchronous (from the caller's point of view) operation that is used to
report a workflow state. Note that a query is inherently read only and cannot
affect a workflow state.

### Run ID
A UUID that a Cadence service assigns to each workflow run. If allowed by
a configured policy, you might be able to re-execute a workflow, after it has
closed or failed, with the same *Workflow ID*. Each such re-execution is called
a run. *Run ID* is used to uniquely identify a run even if it shares a *Workflow ID*
with others.

### Signal
An external asynchronous request to a workflow. It can be used to deliver
notifications or updates to a running workflow at any point in its existence.

### Task
The context needed to execute a specific activity or workflow function.
There are two types of tasks: activity tasks and decision tasks (workflow
tasks).

### Task List
A queue that is persisted inside a Cadence service. When a workflow requests
an activity execution, Cadence service creates an activity task and puts it into
a *Task List*. Then a client-side worker that implements the activity receives
the task from the *Task List* and invokes an activity implementation. For this
to work, the *Task List* name that is used to request an activity execution and
the name used to configure a worker must match.

### Task Token
A unique identifier for a Cadence activity.

### Worker
Also known as a *worker service*. A service that hosts the workflow and
activity implementations. The worker polls the Cadence service for tasks, performs
those tasks, and communicates task execution results back to the Cadence service.
Worker services are developed, deployed, and operated by Cadence customers.

### Workflow
A program that orchestrates activities. A *Workflow* has full control over
which activities are executed, and in which order. A *Workflow* must not affect
the external world directly, only through activities. What makes workflow code
a *Workflow* is that its state is preserved by Cadence. Therefore any failure
of a worker process that hosts the workflow code does not affect the workflow
execution. The *Workflow* continues as if these failures did not happen. At the
same time, activities can fail any moment for any reason. Because workflow code
is fully fault tolerant, it is guaranteed to get notifications about activity
failure or timeouts and act accordingly. There is no limit on potential workflow
duration.

### Workflow Execution
An instance of a *Workflow*. The instance can be in the process of executing
or it could have already completed execution.

### Workflow ID
A unique identifier for a *Workflow Execution*. Cadence guarantees the
uniqueness of an ID within a domain. An attempt to start a *Workflow* with a
duplicate ID results in an **already started** error.

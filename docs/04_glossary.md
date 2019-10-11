# Glossary
This glossary contains terms that are used with the Cadence product.

### Activity
A business-level function that implements your application logic such as calling
a service or transcoding a media file. An activity usually implements a single
well-defined action; it can be short or long running. An activity can be implemented
as a synchronous method or fully asynchronously involving multiple processes.
An activity can be retried indefinitely according to the provided exponential retry policy.
If for any reason an activity is not completed within the specified timeout, an error is reported to the workflow and the workflow decides how to handle it. There is no limit on potential activity
duration.

### Activity Task
A task that contains an activity invocation information that is delivered to an [activity worker](#activity-worker) through and an  [activity task list](#activity-task-list). An activity worker upon receiving activity task executes a correponding [activity](#activity)

### Activity Task List
Task list that is used to deliver [activity tasks](#activity-task) to [activity workers](#activity-worker)

### Activity Worker
An object that is executed in the client application and receives [activity tasks](#activity-task) from an  [activity task list](#activity-task-list) it is subscribed to. Once task is received it invokes a correspondent activity.

### Archival
Archival is a feature that automatically moves [histories](#event-history) from persistence to a blobstore after
the workflow retention period. The purpose of archival is to be able to keep histories as long as needed
while not overwhelming the persistence store. There are two reasons you may want
to keep the histories after the retention period has passed:
1. **Compliance:** For legal reasons, histories may need to be stored for a long period of time.
2. **Debugging:** Old histories can still be accessed for debugging.

### CLI
Cadence command-line interface.

### Client Stub
A client-side proxy used to make remote invocations to an entity that it
represents. For example, to start a workflow, a stub object that represents
this workflow is created through a special API. Then this stub is used to start,
query, or signal the corresponding workflow.

The Go client doesn't use this.

### Decision
Any action taken by the workflow durable function is called a decision. For example:
scheduling an activity, canceling a child workflow, or starting a timer. A [decision task](#decision-task) contains an optional list of decisions. Every decision is recorded in the [event history](#event-history) as an [event](#event).

### Decision Task
Every time a new external event that might affect a workflow state is recorded, a decision task that contains it is added to a [decision-task-list](#decision-task-list) and then picked up by a workflow worker. After the new event is handled, the decision task is completed with a list of [decisions](#decision).
Note that handling of a decision task is usually very fast and is not related to duration
of operations that the workflow invokes.

### Decision Task List
Task list that is used to deliver [decision tasks](#decision-task) to [workflow workers](#workflow-worker)

### Domain
Cadence is backed by a multitenant service. The unit of isolation is called a **domain**. Each domain acts as a namespace for task list names as well as workflow IDs. For example, when a workflow is started, it is started in a
specific domain. Cadence guarantees a unique workflow ID within a domain, and
supports running workflow executions to use the same workflow ID if they are in
different domains. Various configuration options like retention period or archival destination are configured per domain as well through a special CRUD API or through the Cadence CLI. In the multi-cluster deployment, domain is a unit of fail-over. Each domain can only be active on a single Cadence cluster at a time. However, different domains can be active in different clusters and can fail-over independently.

### Event
An indivisible operation performed by your application. For example,
activity_task_started, task_failed, or timer_canceled. Events are recorded in the event history.

### Event History
An append log of events for your application. History is durably persisted
by the Cadence service, enabling seamless recovery of your application state
from crashes or failures. It also serves as an audit log for debugging.

### Local Activity

A [local activity](03_concepts/02_activities#local-activities) is an activity that is invoked directly in the same process by a workflow code. It consumes much less resources than a normal activity, but imposes a lot of limitations like low duration and lack of rate limiting.

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
The context needed to execute a specific activity or workflow state transition.
There are two types of tasks: an [Activity task](#activity-task) and a [Decision task](#decision-task)
(aka workflow task). Note that a single activity execution corresponds to a single activity task,
while a workflow execution employs multiple decision tasks.

### Task List
Common name for [activity task lists](#activity-task-list) and [decision task lists](#decision-task-list)

### Task Token
A unique correlation ID for a Cadence activity. Activity completion calls take either task token
or DomainName, WorkflowID, ActivityID arguments.

### Worker
Also known as a *worker service*. A service that hosts the workflow and
activity implementations. The worker polls the Cadence service for tasks, performs
those tasks, and communicates task execution results back to the Cadence service.
Worker services are developed, deployed, and operated by Cadence customers.

### Workflow
A fault-oblivious stateful function that orchestrates activities. A *Workflow* has full control over
which activities are executed, and in which order. A *Workflow* must not affect
the external world directly, only through activities. What makes workflow code
a *Workflow* is that its state is preserved by Cadence. Therefore any failure
of a worker process that hosts the workflow code does not affect the workflow
execution. The *Workflow* continues as if these failures did not happen. At the
same time, activities can fail any moment for any reason. Because workflow code
is fully fault-oblivious, it is guaranteed to get notifications about activity
failures or timeouts and act accordingly. There is no limit on potential workflow
duration.

### Workflow Execution
An instance of a *Workflow*. The instance can be in the process of executing
or it could have already completed execution.

### Workflow ID
A unique identifier for a *Workflow Execution*. Cadence guarantees the
uniqueness of an ID within a domain. An attempt to start a *Workflow* with a
duplicate ID results in an **already started** error.

### Workflow Task
Synonym of the [Decision Task](#decision-task).

### Workflow Worker
An object that is executed in the client application and receives [decision tasks](#decision-task) from an  [decision task list](#decision-task-list) it is subscribed to. Once task is received it is handled by a correponding workflow.



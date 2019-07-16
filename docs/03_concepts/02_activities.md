# Activities

Fault-oblivious stateful workflow code is the core abstraction of Cadence, but due to deterministic execution requirements they are not allowed to call any external API directly.
Instead they orchestrate execution of activities. In its simplest form a Cadence activity is a function or an object method in one of the supported languages.
Cadence does not recover activity state in case of failures therefore an activity function is allowed to contain any code without restrictions.

Activities are invoked asynchronously though so called task lists. Task list is essentially a queue used to store an activity task until it is picket up by an available worker. The worker processes an activity by invoking its implementation function. When function returns the worker reports the result back to the Cadence service which in turn notifies worklflow about completion. It is possible to implement an activity fully asynchronously by completing it from a different process.

## Timeouts

Cadence does not impose any system limit on activity duration. It is up to application to choose the timeouts for its execution. These are configurable activity timeouts:

- `ScheduleToStart` is a maxiumum time from a workflow requesting activity execution to worker starting its execution. The usual reason for this timeout to fire is all workers being down or not being able to keep up with the request rate. We recommend setting this timeout to the maximum time a workflow is willing to wait for an activity execution in presence of all possible worker outages.
- `StartToClose` is a maximum time an activity can execute after it was picked by a worker.
- `ScheduleToClose` is a maximum time from the workflow requesting an activity execution to its completion.
- `Heartbeat` is a maximum time between heartbeat requests. See [Long Running Activities](#long-running-activities)

Either `ScheduleToClose` or both `ScheduleToStart` and `StartToClose` timeouts are required.

## Retries

As Cadence doesn't recover activities state and they can communicate to any external system failures are expected. Cadence supports automatic activity retries. Any activity when invoked can have an associated retry policy. Here are retry policy parameters:

- `InitialInterval` is a delay before the first retry
- `BackoffCoefficient`. Retry policies are exponential. The coefficient specifies how fast the retry interval is growing. The coefficient of 1 means that retry interval is always equal to the `InitialInterval`.
- `MaximumInterval` specifies the maximum interval between retries. Useful for coefficients more than 1.
- `MaximumAttempts` specifies how many times to attempt to execute an ctivity in presence of failures. If this limit is exceeded the error is returned back to the workflow that invoked the activity. Not required if `ExpirationInterval` is specified.
- `ExpirationInterval` specifies for how long to attempt executing an activity in presence of failures. If this interval is exceeded the error is returned back to the workflow that invoked the activity. Not required if `MaximumAttempts` is specified.
- `NonRetryableErrorReasons` allows to specify erros that shouldn't be retried. For example retrying invalid arguments error doesn't make sense in some scenarios.

There are scenarios when not a single activity but the whole part of a workflow should be retried on failure. For example a media encoding workflow that downloads file to a host, processes it and then uploads result back to storage. In this workflow if host that hosts the worker dies all three activities should be retried on a different host. Such retries should be handled by the workflow code as they are very use case specific.

## Long Running Activities

For long running activities it is recommended to specify a relatively short heartbeat timeout and constantly heartbeat. This way worker failures for even very long running activities can be handled in a timely manner. An activity that specifieds the heartbeat timeout is expected to call heartbeat method _periodically_ from its implementation.

A heartbeat request can include applicaiton specific payload. This is useful to save activity execution progress. If an activity times out due to a missed heartbeat the next attempt to execute it can access that progress and continue its execution from that point.

Long running activities can be used as a special case of leader election. Cadence timeouts use second resolution. So it is not a solution for realtime applications. But if it is OK to react to the process failure within a few seconds then a Cadence heartbeating activity is a good fit.

One common use case for such leader election is monitoring. An activity executes an internal loop that periodically polls some API and checks for some condition. It also heartbeats on every iteration. If condition is satisfied the activity completes which lets its workflow to handle it. If the activity worker dies the activity times out after the heartbeat interval is exceeded and is retried on a different worker. The same pattern works for polling for new files in a S3 buckets or responses in REST or other synchronous APIs.

## Cancellation

A workflow can request an activity cancellation. Currently the only way for an activity to learn that it was cancelled is through heartbeating. The heartbeat request fails with a special error indicating that the activity was cancelled. Then it is up to the activity implementation to perform all the necessary cleanup and report that it is done with it. It is up to the workflow implementation to decide if it wants to wait for the activity cancellation confirmation or just proceed without waiting.

Another common case for activity heartbeat failure is that the workflow that invoked it is being in completed state. In this case an activity is expected to perform cleanup as well.

## Activity Task Routing through Task Lists

Activities are dispatched to workers through task lists. Task lists are queues that workers listen on. Task lists are highly dynamic and lightweight. They don't need to be explicitly registered. And it is OK to have one task list per worker process for example. It is normal to have more than one activity type to be invoked through a single task list. And it is normal in some cases (like host routing) to invoke the same activity type on multiple task lists.

Here are some use cases for employing multiple activity task lists in a single workflow:

- _Flow control_. As task list is a queue a worker that consumes from it asks for an activity task only when it has available capacity. So workers are never overloaded by request spikes. If activity executions are requested faster than workers can process them they are backlogged in the task list.
- _Throttling_. Each activity worker can specify maximum rate it is allowed to processes activities on a task list. And it is not going to exceed this limit even if it has spare capacity. There is also support for global task list rate limiting. This limit works across all workers for the given task list. It frequently used to limit load on some downstream service that activity calls into.
- _Deploying a set of activities independently_. Think about a service that hosts activities and can be deployed indepdently from other activties and workflows. To send activity tasks to this service a separate task list is needed.
- _Workers with different capabilities_. For example workers on GPU boxes vs non GPU boxes. Having two separate task lists in this case allows workflows to pick which one to send activity execution request to.
- _Routing activity to a specific host_. For example in the media encoding case the transform and upload activity have to run on the same host as the download one.
- _Routing activity to a specific process_. For example some activities load large data set and cache it in the process. The activities that rely on this data set should be routed to the same process.
- _Multiple priorities_. One task list per priority and having a worker pool per priority.
- _Versioning_. A new backwards incompatible implementation of an activity might use a different task list.

## Asynchronous Activity Completion

By default an activity is a function or a method depending on a client side library language. As soon as the function returns an activity completes. But in some cases an activity implementation is asynchronous. For example it is forwarded to an external system through a message queue. And the reply comes through a different queue.

To support such use cases Cadence allows activity implementations that do not complete upon activity function completions. A separate API should be used in this case to complete the activity. This API can be called from any process even in a different programming language that the original activity worker used.